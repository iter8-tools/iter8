package action

import (
	"errors"
	"io/ioutil"
	"path"
	"regexp"

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

/*
	containsMetricsTemplate checks if for a metrics template and identifies
	the source of the metrics.

	For example, given "_istio.metrics.tpl", return "istio" or given
	"_ce.metrics.tpl", return "ce". Otherwise, return empty string.
*/
func (gen *GenOpts) getMetricSourceFromTemplate() (string, error) {
	templatesPath := path.Join(gen.chartDir(), "templates")

	fileInfo, err := ioutil.ReadDir(templatesPath)
	if err != nil {
		log.Logger.Error("could not read directory ", templatesPath)
		return "", err
	}

	re := regexp.MustCompile(`^_(\w+)\.metrics\.tpl$`)

	for _, file := range fileInfo {
		fileName := file.Name()

		match := re.FindStringSubmatch(fileName)

		/*
			Given "_istio.metrics.tpl", FindStringSubmatch should return
			["_istio.metrics.tpl", "istio"]
		*/
		if len(match) > 0 {
			if len(match) == 2 {
				return match[1], nil
			} else {
				return "", errors.New("could not properly identify metrics source from " + fileName)
			}
		}
	}

	return "", nil
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

	// check for a metrics template and get its source
	metricSource, err := gen.getMetricSourceFromTemplate()
	if err != nil {
		return err
	}

	// add in metrics.tpl template
	if metricSource != "" {
		mData := []byte(`{{- include "metrics" . }}`)
		c.Templates = append(c.Templates, &chart.File{
			Name: path.Join("templates", driver.ExperimentMetricsPath),
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

	// write metric spec file
	if metricSource != "" {
		metricsBytes := []byte(m[path.Join(c.Name(), "templates", driver.ExperimentMetricsPath)])
		metricsFileName := metricSource + ".metrics.yaml"
		err = ioutil.WriteFile(path.Join(gen.GenDir, metricsFileName), metricsBytes, 0664)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment spec")
			return err
		}
		log.Logger.Infof("created %v file", metricsFileName)
	}

	return err
}
