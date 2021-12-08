package cmd

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	AppHeader               = "APP"
	IdHeader                = "ID"
	CompletedHeader         = "COMPLETED"
	FailedHeader            = "FAILED"
	NumTasksHeader          = "TASKS"
	NumTasksCompletedHeader = "TASKS_COMPLETED"
)

func runGetCmd(cmd *cobra.Command, args []string, o *K8sExperimentOptions) (err error) {
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
	o := newK8sExperimentOptions(streams)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a list of experiments running in the current context",
		Example: `
# Get list of experiments running in cluster
iter8 k get`,
		SilenceUsage: true,
	}
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		// precompute commonly used values derivable from GetOptions
		return o.initK8sExperiment(factory)
		// add any additional precomutation and/or validation here
	}
	cmd.RunE = func(c *cobra.Command, args []string) error {
		return runGetCmd(c, args, o)
	}

	AddExperimentIdOption(cmd, o)
	// Add any other options here

	// Prevent default options from being displayed by the help
	HideGenericCliOptions(cmd)

	return cmd
}
