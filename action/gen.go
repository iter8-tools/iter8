package action

import (
	"io/ioutil"
	"path"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	helmaction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
)

// GenOpts are the options used for generating experiment.yaml
type GenOpts struct {
	// Options provides the values to be combined with the experiment chart
	values.Options
	// SourceDir is the directory containing the Iter8 'charts' folder
	SourceDir string
	// ChartName is the name of the chart
	ChartName string
}

// NewGenOpts initializes and returns gen opts
func NewGenOpts() *GenOpts {
	return &GenOpts{
		SourceDir: ".",
	}
}

// updateChartDependencies for an Iter8 experiment chart
// for now this function has one purpose ...
// bring iter8lib dependency into other experiment charts like load-test-http
func (gen *GenOpts) updateChartDependencies(chartPath string) error {
	// client, settings, cfg are not really initialized with proper values
	// should be ok considering iter8lib is a local file dependency
	client := helmaction.NewDependency()
	settings := cli.New()
	man := &downloader.Manager{
		Out:              ioutil.Discard,
		ChartPath:        chartPath,
		Keyring:          client.Keyring,
		SkipUpdate:       client.SkipRefresh,
		Getters:          getter.All(settings),
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		Debug:            settings.Debug,
	}
	log.Logger.Info("updating chart ", chartPath)
	if err := man.Update(); err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to update chart dependencies")
		return err
	}
	return nil
}

// LocalRun generates a local experiment.yaml file
func (gen *GenOpts) LocalRun() error {
	// chartPath
	chartPath := path.Join(gen.SourceDir, "charts", gen.ChartName)
	if err := gen.updateChartDependencies(chartPath); err != nil {
		return err
	}

	// read in the experiment chart
	c, err := loader.Load(chartPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to load experiment chart")
		return err
	}

	// add in experiment.yaml template
	eData := []byte(`{{- include "experiment" . }}`)
	c.Templates = append(c.Templates, &chart.File{
		Name: path.Join("templates", driver.ExperimentSpecPath),
		Data: eData,
	})

	// get values
	p := getter.All(cli.New())
	v, err := gen.MergeValues(p)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to obtain values for chart")
		return err
	}

	valuesToRender, err := chartutil.ToRenderValues(c, v, chartutil.ReleaseOptions{}, nil)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to compose chart information")
		return err
	}

	// render experiment.yaml
	m, err := engine.Render(c, valuesToRender)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to render chart templates")
		log.Logger.Debug("values: ", valuesToRender)
		return err
	}

	// write experiment spec file
	specBytes := []byte(m[path.Join(c.Name(), "templates", driver.ExperimentSpecPath)])
	err = ioutil.WriteFile(driver.ExperimentSpecPath, specBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment spec")
		return err
	}
	log.Logger.Infof("created %v file", driver.ExperimentSpecPath)

	// build and validate experiment
	fio := &driver.FileDriver{}
	_, err = base.BuildExperiment(false, fio)
	if err != nil {
		return err
	}

	return err
}
