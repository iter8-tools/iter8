package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"

	"github.com/spf13/cobra"
)

var deleteCmd *cobra.Command

func getObjectManifests() ([]string, error) {
	// use function of iter8 gen k8s to create manifest
	generatedManifest, err := basecli.Generate([]string{fmt.Sprintf("%s=%s", "id", k8sExperimentOptions.experimentId)})
	if err != nil {
		return []string{}, err
	}
	// split into objects (as strings)
	objectManifests := strings.Split(generatedManifest.String(), "---")

	return objectManifests, nil
}

func deleteObject(dr dynamic.ResourceInterface, kind string, name string) (err error) {
	deletePolicy := metav1.DeletePropagationForeground
	var gracePeriod int64 = 0
	err = dr.Delete(context.Background(), name, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod, PropagationPolicy: &deletePolicy})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			// don't report error if didn't exist in the first place
			log.Logger.Errorf("delete failed: %s\n", err.Error())
		}
	}
	log.Logger.Info(fmt.Sprintf("deleted %s/%s\n", strings.ToLower(kind), name))

	return nil
}

func deleteObjects() (err error) {
	manifests, err := getObjectManifests()
	if err != nil {
		return err
	}

	// get rest.Config
	restConfig, err := k8sExperimentOptions.ConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	// get RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// prepare dynamic client
	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	for _, manifest := range manifests {
		// convert manifest (string) to obj (unstructured) and get GVK
		obj := &unstructured.Unstructured{}
		_, gvk, err := decoder.Decode([]byte(manifest), nil, obj)
		if err != nil {
			return err
		}

		// find GVR from GVK
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		// obtain REST interface for the GVR
		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources
			dr = dyn.Resource(mapping.Resource).Namespace(k8sExperimentOptions.namespace)
		} else {
			// cluster-wide resources; should not happen for iter8 use cases, but just to be sure
			dr = dyn.Resource(mapping.Resource)
		}

		// do the delete; quit at first error (some objects may remain)
		err = deleteObject(dr, gvk.Kind, obj.GetName())
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	// initialize deleteCmd
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an experiment running in a Kubernetes cluster",
		Example: `
# Delete experiment most recently started in Kubernetes cluster
iter8 k delete

# Delete experient with identifier $EXPERIMENT_ID
iter8 k delete -e $EXPERIMENT_ID`,
		RunE: func(c *cobra.Command, args []string) error {
			k8sExperimentOptions.initK8sExperiment(true)
			log.Logger.Infof("deleting experiment: %s\n", k8sExperimentOptions.experimentId)
			return deleteObjects()
		},
	}
	k8sExperimentOptions.addExperimentIdOption(deleteCmd.Flags())

	// deleteCmd is now initialized
	kCmd.AddCommand(deleteCmd)
}
