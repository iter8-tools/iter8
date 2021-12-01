package get

import (
	"fmt"

	"github.com/iter8-tools/iter8/base/log"

	basecli "github.com/iter8-tools/iter8/cmd"
	"github.com/iter8-tools/iter8/k8s/pkg/utils"

	"github.com/spf13/cobra"
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

	if len(o.experiment) == 0 {
		s, err := utils.GetExperiment(o.client, o.namespace, o.experiment)
		if err != nil {
			return err
		}
		o.experiment = s.GetName()
	}

	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *Options) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *Options) run(cmd *cobra.Command, args []string) (err error) {
	experiments, err := utils.GetExperiments(o.client, o.namespace)
	if err != nil {
		return err
	}

	if len(experiments) == 0 {
		fmt.Println("no experiments found")
		return err
	}

	// fmt.Printf("%-16s  %-9s  %-6s\n", "NAME", "COMPLETED", "FAILED")
	fmt.Printf("%-16s  %-9s  %-6s  %-9s  %-20s\n", "NAME", "COMPLETED", "FAILED", "NUM TASKS", "NUM TASKS COMPLETED")
	for _, experiment := range experiments {

		expIO := &utils.KubernetesExpIO{
			Client:    o.client,
			Namespace: o.namespace,
			Name:      experiment.Name,
		}

		log.Logger.Trace("build started")
		exp, err := basecli.Build(true, expIO)
		log.Logger.Trace("build finished")
		if err != nil {
			return err
		}

		// fmt.Printf("%-16s  %-9t  %-6t\n", experiment.GetName(), exp.Completed(), !exp.NoFailure())
		fmt.Printf("%-16s  %-9t  %-6t  %-9d  %-20d\n", experiment.GetName(), exp.Completed(), !exp.NoFailure(), len(exp.Tasks), exp.Result.NumCompletedTasks)
	}
	return nil
}
