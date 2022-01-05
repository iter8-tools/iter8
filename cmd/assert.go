package cmd

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var assertCmd *cobra.Command

func init() {
	// initialize assertCmd
	assertCmd = basecli.NewAssertCmd()
	assertCmd.Short = `Assert if experiment result satisfies the specified conditions`
	assertCmd.Example = `
# Assert that the most recent experiment has completed
# without failure and its SLOs were satisfied for all versions
iter8 k assert -c completed -c nofailure -c slos

# Another way to express the same assertion
iter8 k assert -c completed,nofailure,slos

# Make assertion about the most recent experiment with app label $APP
iter8 k assert -a $APP -c completed,nofailure,slos

# For experiments with multiple versions, specify that the SLOs for one version were satisfied
iter8 k assert -c completed,nofailure,slosby=0

# The above assertion for an experiment with identifier $ID
iter8 k assert --id $ID -c completed,nofailure,slosby=0

# The above assertion with a runtime timeout
iter8 k assert --id $ID -c completed,nofailure,slosby=0 -t 5s`
	assertCmd.RunE = func(c *cobra.Command, args []string) error {
		k8sExperimentOptions.initK8sExperiment(true)
		log.Logger.Infof("evaluating assert for experiment: %s\n", *k8sExperimentOptions.id)
		allGood, err := k8sExperimentOptions.experiment.Assert(basecli.AssertOptions.Conds, basecli.AssertOptions.Timeout)
		if err != nil || !allGood {
			return err
		}

		if !allGood {
			return errors.New("assert conditions failed")
		}

		return nil
	}

	// initialize options for assertCmd
	assertCmd.Flags().AddFlag(pflag.Lookup("id"))
	assertCmd.Flags().AddFlag(pflag.Lookup("app"))

	// assertCmd is now initialized
	kCmd.AddCommand(assertCmd)
}
