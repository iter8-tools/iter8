package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

const (
	AppHeader               = "APP"
	IdHeader                = "ID"
	CompletedHeader         = "COMPLETED"
	FailedHeader            = "FAILED"
	NumTasksHeader          = "TASKS"
	NumTasksCompletedHeader = "TASKS_COMPLETED"
)

var getCmd *cobra.Command

func runGetCmd(cmd *cobra.Command, args []string, o *K8sExperimentOptions) (err error) {
	experimentSecrets, err := GetExperimentSecrets(o.client, o.namespace, *o.app)
	if err != nil {
		return err
	}

	if len(experimentSecrets) == 0 {
		return errors.New("no experiments found")
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

		exp, err := basecli.Build(true, expIO)
		if err != nil {
			return err
		}

		app := experimentSecret.Labels[AppLabel]
		id := experimentSecret.Labels[IdLabel]
		fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%d\t%d\n", app, id, exp.Completed(), !exp.NoFailure(), len(exp.Tasks), exp.Result.NumCompletedTasks)
	}

	w.Flush()
	fmt.Printf("%s", b.String())
	return nil
}

func init() {
	// initialize getCmd
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get the list of experiments running in Kubernetes",
		Example: `
# Get the list of experiments running in Kubernetes
iter8 k get

# Get list of experiments with app label $APP
iter8 k get -a $APP`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			k8sExperimentOptions.initK8sExperiment(true)
			return runGetCmd(c, args, k8sExperimentOptions)
		},
	}
	// initialize options for getCmd
	getCmd.Flags().AddFlag(basecli.GetIdFlag())
	getCmd.Flags().AddFlag(basecli.GetAppFlag())

	// getCmd is now initialized
	kCmd.AddCommand(getCmd)
}
