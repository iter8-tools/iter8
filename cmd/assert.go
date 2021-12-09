package cmd

import (
	"errors"
	"fmt"

	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
)

var assertCmd *cobra.Command

func init() {
	// initialize assertCmd
	assertCmd = basecli.NewAssertCmd()
	var example = `
# assert that the most recent experiment running in the Kubernetes context is complete
iter8 k assert -c completed`
	assertCmd.Example = fmt.Sprintf("%s%s\n", assertCmd.Example, example)
	assertCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment()
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
