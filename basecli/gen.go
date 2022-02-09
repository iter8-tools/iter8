package basecli

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
)

// GenOptions are the options used by the gen subcommands.
// They store values that can be combined with templates for generating experiment.yaml files Kubernetes manifests.
var GenOptions = values.Options{}

// chartPath	path to experimeent chart folder
var chartPath string

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Render experiment.yaml file by combining an experiment chart with values.",
	Long: `
Render experiment.yaml file by combining an experiment chart with values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// read in the experiment chart
		c, err := loader.Load(chartPath)
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
		fio := &FileExpIO{}
		_, err = Build(false, fio)
		if err != nil {
			return err
		}

		return err
	},
}

func addGenOptions(f *pflag.FlagSet) {
	// See: https://github.com/helm/helm/blob/663a896f4a815053445eec4153677ddc24a0a361/cmd/helm/flags.go#L42 which is the source of these flags
	f.StringSliceVarP(&GenOptions.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&GenOptions.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&GenOptions.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&GenOptions.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
}

func init() {
	genCmd.Flags().StringVarP(&chartPath, "chartPath", "c", "", "path to experiment chart folder")
	genCmd.MarkFlagRequired("chartPath")
	addGenOptions(genCmd.Flags())
	RootCmd.AddCommand(genCmd)
}
