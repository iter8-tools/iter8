package deleter

import (
	"context"
	"fmt"

	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// complete sets all information needed for processing the command
func (o *Options) complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) (err error) {
	o.namespace, _, err = factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.client, err = utils.GetClient(o.ConfigFlags)
	if err != nil {
		return err
	}

	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *Options) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *Options) run(cmd *cobra.Command, args []string) (err error) {
	experiment, err := utils.GetExperiment(o.client, o.namespace, o.experiment)
	if err != nil {
		return err
	}

	o.delete(experiment.Name)

	return nil
}

func (o *Options) delete(e string) {
	fmt.Printf("deleting experiment: %s\n", e)

	ctx := context.Background()
	propPolicy := metav1.DeletePropagationBackground
	options := metav1.DeleteOptions{PropagationPolicy: &propPolicy}
	o.client.CoreV1().Secrets(o.namespace).Delete(ctx, e, options)
	o.client.BatchV1().Jobs(o.namespace).Delete(ctx, e, options)
	o.client.RbacV1().Roles(o.namespace).Delete(ctx, e, options)
	o.client.RbacV1().RoleBindings(o.namespace).Delete(ctx, e, options)
	o.client.CoreV1().Secrets(o.namespace).Delete(ctx, e+"-result", options)
}
