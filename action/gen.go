package action

import (
	"fmt"
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
		Name: path.Join("templates", driver.ExperimentSpecPath),
		Data: eData,
	})

	/*
		attempt to extract providers from valuesToRender

		if providers are available, create the respective metrics files from the
		templates
	*/
	var providers []string
	if rawProviders, ok := c.Values["providers"]; ok {
		// convert iterface to interface array
		convertedRawProviders := rawProviders.([]interface{})

		providers = make([]string, len(convertedRawProviders))
		for i, v := range convertedRawProviders {
			providers[i] = fmt.Sprint(v)
		}
	}

	// add in metrics.yaml template
	for _, provider := range providers {
		// NOTE: This pattern must be documented
		mData := []byte(`{{- include "metrics.` + provider + `" . }}`)
		c.Templates = append(c.Templates, &chart.File{
			Name: path.Join("templates", provider, driver.ExperimentMetricsPath),
			Data: mData,
		})
	}

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

	// write metric spec files
	for _, provider := range providers {
		metricsBytes := []byte(m[path.Join(c.Name(), "templates", provider, driver.ExperimentMetricsPath)])
		metricsFileName := provider + ".metrics.yaml"
		err = ioutil.WriteFile(path.Join(gen.GenDir, metricsFileName), metricsBytes, 0664)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment spec")
			return err
		}
		log.Logger.Infof("created %v file", metricsFileName)
	}

	return err
}
