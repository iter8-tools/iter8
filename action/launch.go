package action

import (
	"strings"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	// DefaultHelmRepository is the URL of the default Helm repository
	DefaultHelmRepository = "https://iter8-tools.github.io/iter8"
	// DefaultChartName is the default name of the Iter8 chart
	DefaultChartName = "iter8"
)

// LaunchOpts are the options used for launching experiments
type LaunchOpts struct {
	// DryRun enables simulating a launch
	DryRun bool
	// ChartPathOptions
	action.ChartPathOptions
	// ChartName is the name of the chart
	ChartName string
	// Options provides the values to be combined with the experiment chart
	values.Options
	// Rundir is the directory where experiment.yaml file is located
	RunDir string
	// KubeDriver enables Kubernetes experiment run
	*driver.KubeDriver
	// LocalChart indicates the chart is on the local filesystem
	LocalChart bool
}

// NewLaunchOpts initializes and returns launch opts
func NewLaunchOpts(kd *driver.KubeDriver) *LaunchOpts {
	return &LaunchOpts{
		DryRun: false,
		ChartPathOptions: action.ChartPathOptions{
			RepoURL: DefaultHelmRepository,
			Version: defaultChartVersion(),
		},
		ChartName:  DefaultChartName,
		Options:    values.Options{},
		RunDir:     ".",
		KubeDriver: kd,
		LocalChart: false,
	}
}

func defaultChartVersion() string {
	return strings.Replace(base.MajorMinor, "v", "", 1) + ".x"
}

// KubeRun launches a Kubernetes experiment
func (lOpts *LaunchOpts) KubeRun() error {
	// initialize kube driver
	if err := lOpts.KubeDriver.Init(); err != nil {
		return err
	}

	return lOpts.KubeDriver.Launch(lOpts.ChartPathOptions, lOpts.ChartName, lOpts.Options, lOpts.Group, lOpts.DryRun)
}
