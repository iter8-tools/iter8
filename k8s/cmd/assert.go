package cmd

import (
	"errors"
	"fmt"

	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/utils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type AssertOptions struct {
	Streams              genericclioptions.IOStreams
	ConfigFlags          *genericclioptions.ConfigFlags
	ResourceBuilderFlags *genericclioptions.ResourceBuilderFlags
	namespace            string
	client               *kubernetes.Clientset

	experimentId string
}

func newAssertOptions(streams genericclioptions.IOStreams) *AssertOptions {
	rbFlags := &genericclioptions.ResourceBuilderFlags{}
	rbFlags.WithAllNamespaces(false)

	return &AssertOptions{
		Streams:              streams,
		ConfigFlags:          genericclioptions.NewConfigFlags(true),
		ResourceBuilderFlags: rbFlags,
	}
}

// complete sets all information needed for processing the command
func (o *AssertOptions) complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) (err error) {

	o.namespace, _, err = factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.client, err = utils.GetClient(o.ConfigFlags)
	if err != nil {
		return err
	}

	if len(o.experimentId) == 0 {
		s, err := utils.GetExperimentSecret(o.client, o.namespace, o.experimentId)
		if err != nil {
			return err
		}
		o.experimentId = s.Labels[utils.IdLabel]
	}

	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *AssertOptions) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *AssertOptions) run(cmd *cobra.Command, args []string) (err error) {
	expIO := &utils.KubernetesExpIO{
		Client:    o.client,
		Namespace: o.namespace,
		Name:      utils.SpecSecretPrefix + o.experimentId,
	}

	log.Logger.Trace("build started")
	exp, err := basecli.Build(true, expIO)
	log.Logger.Trace("build finished")
	if err != nil {
		return err
	}

	allGood, err := exp.Assert(basecli.AssertOptions.Conds, basecli.AssertOptions.Timeout)
	if err != nil || !allGood {
		return err
	}

	if !allGood {
		return errors.New("")
	}

	return nil
}

func NewAssertCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := newAssertOptions(streams)

	cmd := basecli.NewAssertCmd()
	var example = `
# assert that the most recent experiment running in the Kubernetes context is complete
iter8 k assert -c completed`
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if err := o.complete(factory, c, args); err != nil {
			return err
		}
		if err := o.validate(c, args); err != nil {
			return err
		}
		if err := o.run(c, args); err != nil {
			return err
		}
		return nil
	}

	cmd.Flags().StringVarP(&o.experimentId, "experiment-id", "e", "", "remote experiment identifier; if not specified, the most recent experiment is used")

	// Prevent default options from being displayed by the help
	utils.HideGenericCliOptions(cmd)

	return cmd
}
