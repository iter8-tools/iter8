package basecli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const (
	Completed    = "completed"
	NoFailure    = "nofailure"
	SLOs         = "slos"
	SLOsByPrefix = "slosby"
)

// how long to sleep in between retries of asserts
var sleepTime, _ = time.ParseDuration("3s")

// how long have we spent so far in assert attempts
var timeSpent, _ = time.ParseDuration("0s")

type AssertOptionsType struct {
	// assert conditions
	Conds []string
	// timeout for assert conditions to be satisfied
	Timeout time.Duration
}

var AssertOptions = AssertOptionsType{}

// AssertCmd represents the assert command
var AssertCmd = &cobra.Command{
	Use:   "assert",
	Short: "Assert if experiment result satisfies the specified conditions",
	Long: `
Assert if experiment result satisfies the specified conditions. 
If assert conditions are satisfied, exit with code 0. 
Else, return with code 1.`,
	Example: `
# assert that the experiment completed without failures, 
# and SLOs were satisfied
iter8 assert -c completed -c nofailure -c slos

# another way to write the above assertion
iter8 assert -c completed,nofailure,slos

# if the experiment involves multiple app versions, 
# SLOs can be asserted for individual versions
# for example, the following command asserts that
# SLOs are satisfied by version numbered 0
iter8 assert -c completed,nofailures,slosby=0

# timeouts are useful for an experiment that may be long running
# and may run in the background
iter8 assert -c completed,nofailures,slosby=0 -t 5s
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// build experiment
		exp := &Experiment{
			Experiment: &base.Experiment{},
		}
		log.Logger.Trace("build started")
		// replace FileExpIO with ClusterExpIO to build from cluster
		fio := &FileExpIO{}
		exp, err := Build(true, fio)
		log.Logger.Trace("build finished")
		if err != nil {
			return err
		}

		allGood, err := exp.Assert(AssertOptions.Conds, AssertOptions.Timeout)
		if err != nil {
			return err
		}
		if !allGood {
			log.Logger.Error("assert conditions failed")
			os.Exit(1)
		}
		return nil
	},
}

// Assert if experiment satisfies conditions
func (exp *Experiment) Assert(conditions []string, to time.Duration) (bool, error) {
	// check assert conditions
	allGood := true
	for {
		for _, cond := range conditions {
			if strings.ToLower(cond) == Completed {
				c := exp.Completed()
				allGood = allGood && c
				if c {
					log.Logger.Info("experiment completed")
				} else {
					log.Logger.Info("experiment did not complete")
				}
			} else if strings.ToLower(cond) == NoFailure {
				nf := exp.NoFailure()
				allGood = allGood && nf
				if nf {
					log.Logger.Info("experiment has no failure")
				} else {
					log.Logger.Info("experiment failed")
				}
			} else if strings.ToLower(cond) == SLOs {
				slos := exp.SLOs()
				allGood = allGood && slos
				if slos {
					log.Logger.Info("SLOs are satisfied")
				} else {
					log.Logger.Info("SLOs are not satisfied")
				}
			} else if strings.HasPrefix(cond, SLOsByPrefix) {
				version, err := exp.extractVersion(cond)
				if err != nil {
					return false, err
				}
				iv := exp.SLOsBy(version)
				allGood = allGood && iv
				if iv {
					log.Logger.Info("version ", version, " satisfies objectives")
				} else {
					log.Logger.Info("version ", version, " does not satisfy objectives")
				}
			} else {
				log.Logger.Error("unsupported assert condition detected; ", cond)
				return false, fmt.Errorf("unsupported assert condition detected; %v", cond)
			}
		}
		if allGood {
			log.Logger.Info("all conditions were satisfied")
			return true, nil
		} else {
			if timeSpent >= to {
				log.Logger.Info("not all conditions were satisfied")
				return false, nil
			} else {
				log.Logger.Infof("sleeping %v ................................", sleepTime)
				time.Sleep(sleepTime)
				timeSpent += sleepTime
			}
		}
	}
}

// extract version from string
func (exp *Experiment) extractVersion(cond string) (int, error) {
	tokens := strings.Split(cond, "=")
	if len(tokens) != 2 {
		log.Logger.Error("unsupported condition detected; ", cond)
		return -1, fmt.Errorf("unsupported condition detected; %v", cond)
	}
	if exp.Result == nil || exp.Result.Insights == nil {
		log.Logger.Error("insights uninitialized")
		return -1, errors.New("insights is uninitialized")
	}
	for i := 0; i < exp.Result.Insights.NumVersions; i++ {
		if tokens[1] == fmt.Sprintf("%v", i) {
			return i, nil
		}
	}
	log.Logger.Error("number of app versions: ", exp.Result.Insights.NumVersions, "; valid app version must be in the range 0 to ", exp.Result.Insights.NumVersions-1)
	return -1, errors.New(fmt.Sprint("number of app versions: ", exp.Result.Insights.NumVersions, "; valid app version must be in the range 0 to ", exp.Result.Insights.NumVersions-1))
}

func init() {

	AssertCmd.Flags().StringSliceVarP(&AssertOptions.Conds, `condition(s); can specify multiple or separate conditions with commas;`, "c", nil, fmt.Sprintf("%v | %v | %v | %v=<version number>", Completed, NoFailure, SLOs, SLOsByPrefix))
	AssertCmd.MarkFlagRequired("condition")
	AssertCmd.Flags().DurationVarP(&AssertOptions.Timeout, "timeout", "t", 0, "timeout duration (e.g., 5s)")

	RootCmd.AddCommand(AssertCmd)
}
