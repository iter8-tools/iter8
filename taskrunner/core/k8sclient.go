package core

import (
	"context"
	"errors"
	"time"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Much of this k8sclient code is based on the following tutorial:
// https://itnext.io/how-to-generate-client-codes-for-kubernetes-custom-resource-definitions-crd-b4b9907769ba

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return iter8.GroupVersion.WithResource(resource).GroupResource()
}

// GetConfig variable is useful for test mocks.
var GetConfig = func() (*rest.Config, error) {
	return config.GetConfig()
}

// NumAttempt is the number of times to attempt Get operation for a k8s resource
var NumAttempt = 10

// Period is the time duration between between each attempt
var Period = 18 * time.Second

// GetClient constructs and returns a K8s client.
// The returned client has experiment.Experiment type registered.
var GetClient = func() (rc client.Client, err error) {
	var restConf *rest.Config
	restConf, err = GetConfig()
	if err != nil {
		return nil, err
	}

	var addKnownTypes = func(scheme *runtime.Scheme) error {
		// register iter8.GroupVersion and type
		metav1.AddToGroupVersion(scheme, iter8.GroupVersion)
		scheme.AddKnownTypes(iter8.GroupVersion, &Experiment{})

		// Support for notification library
		gv := schema.GroupVersion{
			Group:   "",
			Version: "v1",
		}
		metav1.AddToGroupVersion(scheme, gv)
		scheme.AddKnownTypes(gv, &corev1.Secret{})

		return nil
	}

	var schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	scheme := runtime.NewScheme()
	err = schemeBuilder.AddToScheme(scheme)

	if err == nil {
		rc, err = client.New(restConf, client.Options{
			Scheme: scheme,
		})
		if err == nil {
			return rc, nil
		}
	}
	return nil, errors.New("cannot get client using rest config")
}

// FromCluster fetches an experiment from k8s cluster.
func (b *Builder) FromCluster(nn *client.ObjectKey) *Builder {
	// get the exp; this is a handler (enhanced) exp -- not just an iter8 exp.
	exp := &Experiment{
		Experiment: *iter8.NewExperiment(nn.Name, nn.Namespace).Build(),
	}
	var err error
	if err = GetTypedObject(nn, exp); err == nil {
		b.exp = exp
		return b
	}
	log.Error(err)
	b.err = err
	return b
}

// GetTypedObject gets a typed object from the k8s cluster. Types of such objects include experiment, knative service, etc. This function attempts to get the object `numAttempts` times, with the interval between attempts equal to `period`.
func GetTypedObject(nn *client.ObjectKey, obj client.Object) error {
	log.Trace("Getting typed object: ", obj.GetObjectKind())
	var err error
	var rc client.Client
	if rc, err = GetClient(); err == nil {
		for i := 0; i < NumAttempt; i++ {
			err = rc.Get(context.Background(), *nn, obj)
			if err == nil {
				break
			}
			log.Trace("Finished attempt: ", i, " Total attempts: ", NumAttempt)
			time.Sleep(Period)
			log.Trace("Sleeping period: ", Period)
		}
	}
	return err
}

// UpdateInClusterExperiment updates the experiment within cluster.
func UpdateInClusterExperiment(e *Experiment) (err error) {
	var c client.Client
	if c, err = GetClient(); err == nil {
		err = c.Update(context.Background(), e)
	}
	return err
}

// UpdateInClusterExperimentStatus updates the experiment status within cluster.
func UpdateInClusterExperimentStatus(e *Experiment) (err error) {
	var c client.Client
	if c, err = GetClient(); err == nil {
		err = c.Status().Update(context.Background(), e)
	}
	return err
}
