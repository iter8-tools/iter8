package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/controllers"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/iter8-tools/etc3/taskrunner/tasks/bash"
	"github.com/iter8-tools/etc3/taskrunner/tasks/collect"
	"github.com/iter8-tools/etc3/taskrunner/tasks/exec"
	"github.com/iter8-tools/etc3/taskrunner/tasks/ghaction"
	"github.com/iter8-tools/etc3/taskrunner/tasks/http"
	"github.com/iter8-tools/etc3/taskrunner/tasks/readiness"
	"github.com/iter8-tools/etc3/taskrunner/tasks/runscript"
	"github.com/iter8-tools/etc3/taskrunner/tasks/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/types"
)

// getExperimentNN gets the name and namespace of the experiment from environment variables.
// Returns error if unsuccessful.
func getExperimentNN() (*types.NamespacedName, error) {
	name := viper.GetViper().GetString("experiment_name")
	namespace := viper.GetViper().GetString("experiment_namespace")
	if len(name) == 0 || len(namespace) == 0 {
		return nil, errors.New("invalid experiment name/namespace")
	}
	return &types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, nil
}

// GetAction converts an action spec into an action.
func GetAction(exp *core.Experiment, actionSpec v2alpha2.Action) (core.Action, error) {
	actionSlice := make(core.Action, len(actionSpec))
	var err error
	for i := 0; i < len(actionSpec); i++ {
		// if this is a run ... populate runspec
		if core.IsARun(&actionSpec[i]) {
			if actionSlice[i], err = runscript.Make(&actionSpec[i]); err != nil {
				break
			}
		} else if core.IsATask(&actionSpec[i]) {
			if actionSlice[i], err = MakeTask(&actionSpec[i]); err != nil {
				break
			}
		} else {
			return nil, errors.New("action spec contains item that is neither run spec nor task spec")
		}
	}
	return actionSlice, err
}

// run is a helper function used in the definition of runCmd cobra command.
func run(cmd *cobra.Command, args []string) error {
	nn, err := getExperimentNN()
	var exp *core.Experiment
	if err == nil {
		if exp, err = (&core.Builder{}).FromCluster(nn).Build(); err == nil {
			var actionSpec v2alpha2.Action
			if actionSpec, err = exp.GetActionSpec(action); err == nil {
				var actionSlice core.Action
				if actionSlice, err = GetAction(exp, actionSpec); err == nil {
					ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)
					ctx = context.WithValue(ctx, core.ContextKey("action"), action)
					// pass in the type of action within context ...
					log.Trace("created context for experiment")
					err = actionSlice.Run(ctx)
					if err == nil {
						return nil
					}
				}
			} else {
				log.Error("could not find specified action: " + action)
				return nil
			}
		}
	}

	if err != nil {
		il := controllers.Iter8Log{
			IsIter8Log:          true,
			ExperimentName:      exp.Name,
			ExperimentNamespace: exp.Namespace,
			Source:              controllers.Iter8LogSourceTR,
			Priority:            controllers.Iter8LogPriorityHigh,
			Message:             err.Error(),
			Precedence:          core.GetIter8LogPrecedence(exp, action),
		}
		fmt.Println(il.JSON())
	}

	return err
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run an action",
	Long:  `Sequentially execute all tasks in the specified action; if any task run results in an error, exit immediately with error.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			log.Error("Exiting with error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVarP(&action, "action", "a", "", "name of the action")
	runCmd.MarkPersistentFlagRequired("action")
}

// MakeTask constructs a Task from a TaskSpec or returns an error if any.
func MakeTask(t *v2alpha2.TaskSpec) (core.Task, error) {
	if t == nil || t.Task == nil || len(*t.Task) == 0 {
		return nil, errors.New("nil or empty task found")
	}
	switch *t.Task {
	case bash.TaskName:
		return bash.Make(t)
	case collect.TaskName:
		return collect.Make(t)
	case exec.TaskName:
		return exec.Make(t)
	case ghaction.TaskName:
		return ghaction.Make(t)
	case http.TaskName:
		return http.Make(t)
	case readiness.TaskName:
		return readiness.Make(t)
	case slack.TaskName:
		return slack.Make(t)
	default:
		return nil, errors.New("unknown task: " + *t.Task)
	}
}
