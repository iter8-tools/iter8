package action

import (
	"helm.sh/helm/v3/pkg/action"
)

type ChartNameAndDestOptions struct {
	ChartName string
	DestDir   string
}

type Launch struct {
	DryRun bool
	ChartNameAndDestOptions
	ExperimentGroup
	action.Install
}

func NewLaunch(cfg *action.Configuration) *Launch {
	return &Launch{}
}
