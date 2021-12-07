package cmd

import (
	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type RunOptions struct {
	Streams              genericclioptions.IOStreams
	ConfigFlags          *genericclioptions.ConfigFlags
	ResourceBuilderFlags *genericclioptions.ResourceBuilderFlags
	namespace            string
	client               *kubernetes.Clientset

	experimentId string
}

func newRunOptions(streams genericclioptions.IOStreams) *RunOptions {
	rbFlags := &genericclioptions.ResourceBuilderFlags{}
	rbFlags.WithAllNamespaces(false)

	return &RunOptions{
		Streams:              streams,
		ConfigFlags:          genericclioptions.NewConfigFlags(true),
		ResourceBuilderFlags: rbFlags,
	}
}

// complete sets all information needed for processing the command
func (o *RunOptions) complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) (err error) {
	log.Logger.Trace("iter8 run complete() called")
	defer log.Logger.Trace("iter8 run complete() completed")

	o.namespace, _, err = factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.client, err = GetClient(o.ConfigFlags)
	if err != nil {
		return err
	}

	if len(o.experimentId) == 0 {
		s, err := GetExperimentSecret(o.client, o.namespace, o.experimentId)
		if err != nil {
			return err
		}
		o.experimentId = s.Labels[IdLabel]
	}

	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *RunOptions) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *RunOptions) run(cmd *cobra.Command, args []string) (err error) {
	log.Logger.Trace("iter8 run run() called")
	defer log.Logger.Trace("iter8 run run() completed")

	expIO := &KubernetesExpIO{
		Client:    o.client,
		Namespace: o.namespace,
		Name:      SpecSecretPrefix + o.experimentId,
	}

	log.Logger.Trace("iter8 run: build started")
	exp, err := basecli.Build(false, expIO)
	log.Logger.Trace("iter8 run: build finished")
	if err != nil {
		return err
	}

	return exp.Run(expIO)
}

func NewRunCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := basecli.NewRunCmd()
	cmd.Hidden = true

	o := newRunOptions(streams)

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
	HideGenericCliOptions(cmd)

	return cmd
}
