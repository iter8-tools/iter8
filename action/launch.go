package action

import (
	"os"
	"path"
	"path/filepath"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	// DefaultHelmRepository is the URL of the default Helm repository
	DefaultHelmRepository = "https://iter8-tools.github.io/hub"
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
		DryRun:           false,
		ChartPathOptions: action.ChartPathOptions{},
		ChartName:        "",
		Options:          values.Options{},
		RunDir:           ".",
		KubeDriver:       kd,
		LocalChart:       false,
	}
}

// LocalRun launches a local experiment
func (lOpts *LaunchOpts) LocalRun() error {
	log.Logger.Debug("launch local run started...")

	var gOpts GenOpts
	if lOpts.LocalChart {
		// local charts
		gOpts = GenOpts{
			Options:   lOpts.Options,
			GenDir:    lOpts.RunDir,
			ChartName: path.Join(filepath.Dir(lOpts.ChartName), filepath.Base(lOpts.ChartName)),
		}
	} else {
		// non-local charts. download charts

		// create temporary folder to store chart
		chartsFolderName, err := os.MkdirTemp("", "iter8-")
		if err != nil {
			log.Logger.Error("failed to download chart")
			return err
		}
		defer func() {
			// ignore error value
			_ = os.RemoveAll(chartsFolderName)
		}()

		// pull chart (and its dependencies)
		client := action.NewPullWithOpts(action.WithConfig(&action.Configuration{}))
		client.RepoURL = lOpts.RepoURL
		client.Untar = true
		client.UntarDir = chartsFolderName + "/charts"
		client.Settings = cli.New()
		log.Logger.Debug("client.UntarDir ", client.UntarDir)

		_, err = client.Run(lOpts.ChartName)
		if err != nil {
			return err
		}

		log.Logger.Trace("chart pulled")

		// gen experiment spec
		gOpts = GenOpts{
			Options:   lOpts.Options,
			GenDir:    lOpts.RunDir,
			ChartName: path.Join(client.UntarDir, lOpts.ChartName),
		}
	}

	if err := gOpts.LocalRun(); err != nil {
		return err
	}
	log.Logger.Debug("gen complete")

	// all done if this is a dry run
	if lOpts.DryRun {
		log.Logger.Info("dry run complete")
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

	return lOpts.KubeDriver.Launch(lOpts.ChartPathOptions, lOpts.ChartName, lOpts.Options, lOpts.Group, lOpts.DryRun)
}
