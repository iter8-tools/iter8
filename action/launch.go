package action

import (
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/cli/values"
)

// LaunchOpts are the options used for launching experiments
type LaunchOpts struct {
	// DryRun enables simulating a launch
	DryRun bool
	// GitFolder is the full path to the GitHub Iter8 experiment charts folder
	GitFolder string
	// ChartsParentDir is the directory where `charts` is to be downloaded or is located
	ChartsParentDir string
	// NoDownload disables charts download.
	// With this option turned on, `charts` that are already present locally are reused
	NoDownload bool
	// ChartName is the name of the chart
	ChartName string
	// Options provides the values to be combined with the experiment chart
	values.Options
	// Rundir is the directory where experiment.yaml file is located
	RunDir string
	// KubeDriver enables Kubernetes experiment run
	*driver.KubeDriver
}

// NewHubOpts initializes and returns launch opts
func NewLaunchOpts(kd *driver.KubeDriver) *LaunchOpts {
	return &LaunchOpts{
		DryRun:          false,
		GitFolder:       DefaultGitFolder,
		ChartsParentDir: ".",
		NoDownload:      false,
		ChartName:       "",
		Options:         values.Options{},
		RunDir:          ".",
		KubeDriver:      kd,
	}
}

// LocalRun launches a local experiment
func (lOpts *LaunchOpts) LocalRun() error {
	log.Logger.Debug("launch local run started...")
	if !lOpts.NoDownload {
		// download chart from Iter8 hub
		hOpts := &HubOpts{
			GitFolder:       lOpts.GitFolder,
			ChartsParentDir: lOpts.ChartsParentDir,
		}
		if err := hOpts.LocalRun(); err != nil {
			return err
		}
		log.Logger.Debug("hub complete")
	} else {
		log.Logger.Info("using `charts` under ", lOpts.ChartsParentDir)
	}

	// gen experiment spec
	gOpts := GenOpts{
		Options:         lOpts.Options,
		ChartsParentDir: lOpts.ChartsParentDir,
		GenDir:          lOpts.RunDir,
		ChartName:       lOpts.ChartName,
	}
	if err := gOpts.LocalRun(); err != nil {
		return err
	}
	log.Logger.Debug("gen complete")

	// all done if this is a dry run
	if lOpts.DryRun {
		return nil
	}

	// run experiment locally
	log.Logger.Info("starting local experiment")
	rOpts := &RunOpts{
		RunDir:     lOpts.RunDir,
		KubeDriver: lOpts.KubeDriver,
	}
	return rOpts.LocalRun()
}

// KubeRun launches a Kubernetes experiment
func (lOpts *LaunchOpts) KubeRun() error {
	// initialize kube driver
	if err := lOpts.KubeDriver.Init(); err != nil {
		return err
	}
	// update dependencies
	gOpts := GenOpts{
		Options:         lOpts.Options,
		ChartsParentDir: lOpts.ChartsParentDir,
		GenDir:          lOpts.RunDir,
		ChartName:       lOpts.ChartName,
	}
	gOpts.updateChartDependencies()

	if lOpts.Revision > 0 { // last release found; setup upgrade
		return lOpts.KubeDriver.Upgrade(gOpts.chartDir(), lOpts.Options, lOpts.Group, lOpts.DryRun)
	} else {
		// no release found; setup install
		return lOpts.KubeDriver.Install(gOpts.chartDir(), lOpts.Options, lOpts.Group, lOpts.DryRun)
	}
}
