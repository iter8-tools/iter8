package action

import (
	"io/ioutil"
	"path"

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

const chartsFolderName = "charts"

// GenOpts are the options used for generating experiment.yaml
type GenOpts struct {
	// Options provides the values to be combined with the experiment chart
	values.Options
	// ChartsParentDir is the directory where `charts` directory is located
	ChartsParentDir string
	// GenDir is the directory where the chart templates are rendered
	GenDir string
	// ChartName is the name of the chart
	ChartName string
}

// NewGenOpts initializes and returns gen opts
func NewGenOpts() *GenOpts {
	return &GenOpts{
		ChartsParentDir: ".",
		GenDir:          ".",
	}
}

// chartDir returns the path to chart directory
func (gen *GenOpts) chartDir() string {
	return path.Join(gen.ChartsParentDir, chartsFolderName, gen.ChartName)
}

// updateChartDependencies for an Iter8 experiment chart
// for now this function has one purpose ...
// bring iter8lib dependency into other experiment charts like load-test-http
func (gen *GenOpts) updateChartDependencies() error {
	// client, settings, cfg are not really initialized with proper values
	// should be ok considering iter8lib is a local file dependency
	client := helmaction.NewDependency()
	settings := cli.New()
	man := &downloader.Manager{
		Out:              ioutil.Discard,
		ChartPath:        gen.chartDir(),
		Keyring:          client.Keyring,
		SkipUpdate:       client.SkipRefresh,
		Getters:          getter.All(settings),
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		Debug:            settings.Debug,
	}
	log.Logger.Info("updating chart ", gen.chartDir())
	if err := man.Update(); err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to update chart dependencies")
		return err
	}
	return nil
}

// LocalRun generates a local experiment.yaml file
func (gen *GenOpts) LocalRun() error {
	// chartPath
	if err := gen.updateChartDependencies(); err != nil {
		return err
	}

	chartPath := path.Join(gen.ChartsParentDir, chartsFolderName, gen.ChartName)
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
	err = ioutil.WriteFile(path.Join(gen.GenDir, driver.ExperimentSpecPath), specBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment spec")
		return err
	}
	log.Logger.Infof("created %v file", driver.ExperimentSpecPath)

	return err
}
