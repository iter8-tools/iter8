package core

// TaskMeta is common to all Tasks
type TaskMeta struct {
	// Task uniquely identifies the task to be executed.
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	// Run is a special type of task meant to run a bash script.
	// TaskSpec must include exactly one of the two fields, run or task.
	Run *string `json:"run,omitempty" yaml:"run,omitempty"`
	// If specifies if this task should be executed.
	// Task will be evaluated if condition specified by if evaluates to true, and not otherwise.
	If *string `json:"if,omitempty" yaml:"if,omitempty"`
}

// TaskSpec contains the specification of a task.
type TaskSpec struct {
	TaskMeta
	// With holds inputs to this task.
	With map[string]interface{} `json:"with,omitempty" yaml:"with,omitempty"`
}

// Task objects can be run
type Task interface {
	Run(exp *Experiment) error
}

// TaskMaker can make tasks from task specs
type TaskMaker interface {
	Make(ts *TaskSpec) (Task, error)
}
