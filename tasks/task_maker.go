package task

import (
	"errors"

	"github.com/iter8-tools/iter8/core"
)

type TaskMaker struct{}

// Make constructs a Task from a tasp spec; returns error if any
func (tm *TaskMaker) Make(t *core.TaskSpec) (core.Task, error) {
	if t == nil || t.Task == nil || len(*t.Task) == 0 {
		core.Logger.WithStackTrace(t.String()).Error("nil or empty task found")
		return nil, errors.New("nil or empty task found")
	}
	switch *t.Task {
	case CollectTaskName:
		return MakeCollect(t)
	case AssessTaskName:
		return MakeAssess(t)
	default:
		return nil, errors.New("unknown task: " + *t.Task)
	}
}
