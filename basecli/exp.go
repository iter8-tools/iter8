package basecli

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
)

// expCmd represents the exp command
var expCmd = &cobra.Command{
	Use:   "exp",
	Short: "Render experiment.yaml file by combining an experiment chart with values.",
	Long: `
Render experiment.yaml file by combining an experiment chart with values.
This command is intended to be run from the root of an Iter8 experiment chart. Values may be specified and are processed in the same manner as they are for Helm charts.`,
	Example: `
iter8 gen exp --set url=https://httpbin.org/get
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// read in the experiment chart
		c, err := loader.Load(".")
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to load experiment chart")
			return err
		}

		// add in experiment.yaml template
		eData := []byte(fmt.Sprintf(`{{- include "%v.experiment" . }}`, c.Name()))
		c.Templates = append(c.Templates, &chart.File{
			Name: path.Join("templates", experimentSpecPath),
			Data: eData,
		})

		// get values
		p := getter.All(cli.New())
		v, err := GenOptions.MergeValues(p)
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
			log.Logger.Error("values: ", valuesToRender)
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
		fio := &FileExpIO{}
		_, err = Build(false, fio)
		if err != nil {
			return err
		}

		return err
	},
	Hidden: true,
}

func init() {
	genCmd.AddCommand(expCmd)
}
