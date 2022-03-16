package cmd

import (
	"io"
	"os"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
)

var (
	logLevel            = "info"
	settings            = cli.New()
	kd                  = driver.NewKubeDriver(settings)
	outStream io.Writer = os.Stdout
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Kubernetes release optimizer",
	Long: `
Kubernetes release optimizer built for DevOps, MLOps, SRE, and data science teams.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ll, err := logrus.ParseLevel(logLevel)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
		log.Logger.Level = ll
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// disable completion command for now
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "l", "info", "trace, debug, info, warning, error, fatal, panic")
}

// Credits: the following function is from Helm. Please see:
// https://github.com/helm/helm/blob/main/cmd/helm/flags.go
func addValueFlags(f *pflag.FlagSet, v *values.Options) {
	f.StringSliceVarP(&v.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&v.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
}

// Credits: the following function is modified from Helm.
// Please see addChartPathFlags below:
// https://github.com/helm/helm/blob/main/cmd/helm/flags.go
func addChartFlags(cmd *cobra.Command, c *action.ChartPathOptions, nd *ia.ChartNameAndDestOptions) {
	// fill nd
	cmd.Flags().StringVarP(&nd.ChartName, "chartName", "c", "", "name of the experiment chart")
	cmd.MarkFlagRequired("chartName")
	cmd.Flags().StringVar(&nd.DestDir, "destDir", ".", "destination folder where experiment chart is downloaded and unpacked")

	// fill c
	cmd.Flags().StringVar(&c.Version, "version", "", "specify a version constraint for the chart version to use. This constraint can be a specific tag (e.g. 0.9.0) or it may reference a valid range (e.g. 0.9.x). If this is not specified, the latest compatible version is used")
	cmd.Flags().StringVar(&c.RepoURL, "repoURL", driver.DefaultIter8RepoURL, "experiment chart repository url where to locate the requested experiment chart")
}
