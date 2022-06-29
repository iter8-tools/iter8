package action

import (
	"io/ioutil"
	"path"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
)

const (
	ChartsFolderName = "charts"
	DefaultChartName = "iter8"
)

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
		ChartName:       DefaultChartName,
	}
}

// chartDir returns the path to chart directory
func (gen *GenOpts) chartDir() string {
	return path.Join(gen.ChartsParentDir, ChartsFolderName, gen.ChartName)
}

// LocalRun generates a local experiment.yaml file
func (gen *GenOpts) LocalRun() error {
	// update dependencies
	if err := driver.UpdateChartDependencies(gen.chartDir(), nil); err != nil {
		return err
	}

	// read in the experiment chart
	c, err := loader.Load(gen.chartDir())
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to load experiment chart")
		return err
	}

	// add in experiment.yaml template
	eData := []byte(`{{- include "experiment" . }}`)
	c.Templates = append(c.Templates, &chart.File{
		Name: path.Join("templates", driver.ExperimentPath),
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

	// write experiment
	expBytes := []byte(m[path.Join(c.Name(), "templates", driver.ExperimentPath)])
	err = ioutil.WriteFile(path.Join(gen.GenDir, driver.ExperimentPath), expBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment")
		return err
	}
	log.Logger.Infof("created %v file", driver.ExperimentPath)

	return err
}
