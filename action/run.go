package action

import (
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
)

type RunOpts struct {
	RunDir string
	// applicable only for kubernetes experiments
	Group string
}

func NewRunOpts() *RunOpts {
	return &RunOpts{
		RunDir: ".",
	}
}

func (runner *RunOpts) LocalRun() error {
	return base.RunExperiment(&driver.FileDriver{
		RunDir: runner.RunDir,
	})
}

// # ---
// # apiVersion: v1
// # kind: Secret
// # metadata:
// #   name: {{ $name }}-result
// # stringData:
// #   result.yaml: |
// #     startTime: {{ now }}
// #     numCompletedTasks: 0
// #     failure: false
// #     iter8Version: {{ .Chart.AppVersion }}
