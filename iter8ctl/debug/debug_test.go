package debug

import (
	"sort"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/controllers"
	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ils []controllers.Iter8Log = []controllers.Iter8Log{
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityHigh,
		Message:             "hello world",
		Precedence:          3,
	},
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityMedium,
		Message:             "hello world",
		Precedence:          0,
	},
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityMedium,
		Message:             "hello world",
		Precedence:          2,
	},
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityMedium,
		Message:             "hello world",
		Precedence:          1,
	},
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityHigh,
		Message:             "hello world again",
		Precedence:          1,
	},
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityLow,
		Message:             "hello world again and again",
		Precedence:          1,
	},
	{
		IsIter8Log:          true,
		ExperimentName:      "hello",
		ExperimentNamespace: "default",
		Source:              "task-runner",
		Priority:            controllers.Iter8LogPriorityHigh,
		Message:             "hello world",
		Precedence:          4,
	},
}

func TestSortIter8Logs(t *testing.T) {
	sortedIl := []controllers.Iter8Log{ils[1], ils[3], ils[4], ils[5], ils[2], ils[0], ils[6]}

	// sort logs by precedence
	sort.Sort(byPrecedence(ils))

	assert.Equal(t, ils, sortedIl)
}

func TestDebug(t *testing.T) {
	// we will mock this function
	gtrl := getTaskRunnerLogs

	// filter (keep) medium and high priority logs
	filteredIls := []controllers.Iter8Log{}
	for _, il := range ils {
		if il.Priority <= controllers.Iter8LogPriorityMedium {
			filteredIls = append(filteredIls, il)
		}
	}

	// mocked function; will return some non-Iter8log-lines combined with Iter8Log lines
	getTaskRunnerLogs = func(exp *expr.Experiment) ([]byte, error) {
		var tr string
		tr += "tango echo charlie\n"
		for _, il := range ils {
			tr = tr + il.JSON() + "\n"
		}
		tr += "how about we finish up?\n"
		return []byte(tr), nil
	}

	exp := &expr.Experiment{
		Experiment: v2alpha2.Experiment{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hello",
				Namespace: "default",
			},
			Spec:   v2alpha2.ExperimentSpec{},
			Status: v2alpha2.ExperimentStatus{},
		},
	}
	actualIls, err := Debug(exp, controllers.Iter8LogPriorityMedium)

	// unmock
	getTaskRunnerLogs = gtrl

	assert.NoError(t, err)
	assert.Equal(t, filteredIls, actualIls)
}
