package get

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/iter8-tools/iter8/base/log"

	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
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
func (o *Options) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *Options) run(cmd *cobra.Command, args []string) (err error) {
	experimentSecrets, err := utils.GetExperimentSecrets(o.client, o.namespace)
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

		expIO := &utils.KubernetesExpIO{
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

		app := experimentSecret.Labels[utils.AppLabel]
		id := experimentSecret.Labels[utils.IdLabel]
		fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%d\t%d\n", app, id, exp.Completed(), !exp.NoFailure(), len(exp.Tasks), exp.Result.NumCompletedTasks)
		w.Flush()
	}

	fmt.Printf("%s", b.String())
	return nil
}
