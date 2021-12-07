package cmd

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type GetOptions struct {
	Streams              genericclioptions.IOStreams
	ConfigFlags          *genericclioptions.ConfigFlags
	ResourceBuilderFlags *genericclioptions.ResourceBuilderFlags
	namespace            string
	client               *kubernetes.Clientset

	experimentId string
}

func newGetOptions(streams genericclioptions.IOStreams) *GetOptions {
	rbFlags := &genericclioptions.ResourceBuilderFlags{}
	rbFlags.WithAllNamespaces(false)

	return &GetOptions{
		Streams:              streams,
		ConfigFlags:          genericclioptions.NewConfigFlags(true),
		ResourceBuilderFlags: rbFlags,
	}
}

const (
	AppHeader               = "APP"
	IdHeader                = "ID"
	CompletedHeader         = "COMPLETED"
	FailedHeader            = "FAILED"
	NumTasksHeader          = "TASKS"
	NumTasksCompletedHeader = "TASKS_COMPLETED"
)

// complete sets all information needed for processing the command
func (o *GetOptions) complete(factory cmdutil.Factory, cmd *cobra.Command, args []string) (err error) {
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
func (o *GetOptions) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *GetOptions) run(cmd *cobra.Command, args []string) (err error) {
	experimentSecrets, err := GetExperimentSecrets(o.client, o.namespace)
	if err != nil {
		return err
	}

	if len(experimentSecrets) == 0 {
		fmt.Println("no experiments found")
		return err
	}

	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", AppHeader, IdHeader, CompletedHeader, FailedHeader, NumTasksHeader, NumTasksCompletedHeader)
	for _, experimentSecret := range experimentSecrets {

		expIO := &KubernetesExpIO{
			Client:    o.client,
			Namespace: o.namespace,
			Name:      experimentSecret.Name,
		}

		log.Logger.Trace("build started")
		exp, err := basecli.Build(true, expIO)
		log.Logger.Trace("build finished")
		if err != nil {
			return err
		}

		app := experimentSecret.Labels[AppLabel]
		id := experimentSecret.Labels[IdLabel]
		fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%d\t%d\n", app, id, exp.Completed(), !exp.NoFailure(), len(exp.Tasks), exp.Result.NumCompletedTasks)
		w.Flush()
	}

	fmt.Printf("%s", b.String())
	return nil
}

func NewGetCmd(factory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := newGetOptions(streams)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a list of experiments running in the current context",
		Example: `
# Get list of experiments running in cluster
iter8 get`,
		SilenceUsage: true,
	}
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

	cmd.Flags().StringVarP(&o.experimentId, "experiment-id", "e", "", "remote experiment identifier")

	// Prevent default options from being displayed by the help
	HideGenericCliOptions(cmd)

	return cmd
}
