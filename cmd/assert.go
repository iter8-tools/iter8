package cmd

import (
	"errors"
	"fmt"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var assertCmd *cobra.Command

func init() {
	// initialize assertCmd
	assertCmd = basecli.NewAssertCmd()
	var example = `
# assert that the most recent experiment running in a Kubernetes cluster has completed
iter8 k assert -c completed

# assert experient with identifier $EXPERIMENT_ID has completed
iter8 k assert -e $EXPERIMENT_ID -c completed`
	assertCmd.Example = fmt.Sprintf("%s%s\n", assertCmd.Example, example)

	assertCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(true)
		log.Logger.Infof("evaluating assert for experiment: %s\n", k8sExperimentOptions.experimentId)
		allGood, err := k8sExperimentOptions.experiment.Assert(basecli.AssertOptions.Conds, basecli.AssertOptions.Timeout)
		if err != nil || !allGood {
			return err
		}

		if !allGood {
			return errors.New("assert conditions failed")
		}

		return nil
	}
	k8sExperimentOptions.addExperimentIdOption(assertCmd.Flags())

	// assertCmd is now initialized
	kCmd.AddCommand(assertCmd)
}
