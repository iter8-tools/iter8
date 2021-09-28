// Package debug implements the `iter8ctl debug` subcommand.
package debug

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/iter8-tools/etc3/controllers"
	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/iter8-tools/etc3/iter8ctl/utils"
)

const (
	iter8NameSpace       string = "iter8-system"
	iter8ExpNameKey      string = "iter8/experimentName"
	iter8ExpNamespaceKey string = "iter8/experimentNamespace"
)

// byPrecedence implements sort.Interface based on the precedence of Iter8Log
type byPrecedence []controllers.Iter8Log

// Len returns length of the log slice
func (a byPrecedence) Len() int {
	return len(a)
}

// Less is true if i^th log should precede the j^th log and false otherwise
func (a byPrecedence) Less(i, j int) bool {
	if a[i].Source == a[j].Source && a[i].Source == controllers.Iter8LogSourceTR {
		if a[i].Precedence < a[j].Precedence {
			return true
		} else if a[i].Precedence == a[j].Precedence {
			return i < j
		} else {
			return false
		}
	} else {
		panic(fmt.Sprintf("only supported source at the moment is %s", controllers.Iter8LogSourceTR))
	}
}

// Swap two entries in the log slice
func (a byPrecedence) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// getTaskRunnerLogs gets the logs for the task runner jobs for the given experiment
// var useful for mocking in tests
var getTaskRunnerLogs = func(exp *expr.Experiment) ([]byte, error) {
	selector := fmt.Sprintf("%s=%s,%s=%s", iter8ExpNameKey, exp.Name, iter8ExpNamespaceKey, exp.Namespace)

	cmd := exec.Command("kubectl", "logs", "-l", selector, "-n", iter8NameSpace, "--tail=-1")
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	return stdout, nil
}

// Debug prints iter8-logs for the given experiment
func Debug(exp *expr.Experiment, priority controllers.Iter8LogPriority) ([]controllers.Iter8Log, error) {
	// check priority
	if priority < 1 || priority > 3 {
		return nil, errors.New("priority can only be set to 1, 2, or 3")
	}

	// fetch task runner job logs
	tr, err := getTaskRunnerLogs(exp)
	if err != nil {
		return nil, err
	}

	// fetch controller logs
	// fetch analytics logs

	// initialize Iter8logs
	ils := []controllers.Iter8Log{}

	scanner := bufio.NewScanner(strings.NewReader(string(tr)))
	for scanner.Scan() {
		line := scanner.Text()
		if utils.IsJSONObject(line) {
			il := controllers.Iter8Log{}
			if json.Unmarshal([]byte(line), &il) == nil {
				// filter Iter8logs for this experiment
				if il.IsIter8Log &&
					il.ExperimentName == exp.Name &&
					il.ExperimentNamespace == exp.Namespace &&
					il.Priority <= priority {
					ils = append(ils, il)
				}
			}
		}

		// sort logs by precedence
		sort.Sort(byPrecedence(ils))
	}

	// return iter8-logs sorted by precedence
	return ils, nil

}
