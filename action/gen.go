package action

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
)

type Gen struct {
	SourceDir string
	values.Options
}

func NewGen() *Gen {
	return &Gen{}
}

func (gen *Gen) gen() error {
	// read in the experiment chart
	c, err := loader.Load(gen.SourceDir)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to load experiment chart")
		return err
	}

	// check version
	if c.AppVersion() != base.MajorMinor {
		err = fmt.Errorf("chart's app version (%v) and Iter8 CLI version (%v) do not match", c.AppVersion(), base.MajorMinor)
		log.Logger.Error(err)
		return err
	}

	// add in experiment.yaml template
	eData := []byte(`{{- include "experiment" . }}`)
	c.Templates = append(c.Templates, &chart.File{
		Name: path.Join("templates", experimentSpecPath),
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
		log.Logger.WithStackTrace(err.Error()).Error("unable to render chart")
		log.Logger.Debug("values: ", valuesToRender)
		return err
	}

	// write experiment spec file
	specBytes := []byte(m[path.Join(c.Name(), "templates", experimentSpecPath)])
	err = ioutil.WriteFile(experimentSpecPath, specBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment spec")
		return err
	}
	log.Logger.Info("Created the experiment.yaml file containing the experiment spec")

	// build and validate experiment
	fio := &fileOps{}
	_, err = build(false, fio)
	if err != nil {
		return err
	}

	return err

}
