package cmd

import (
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli"
)

var (
	// default log level for Iter8 CLI
	logLevel = "info"
	// Default Helm and Kubernetes settings
	settings = cli.New()
	// KubeDriver used by actions package
	kd = driver.NewKubeDriver(settings)
	// kubeclient is the client used for controllers package
	kubeClient k8sclient.Interface
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iter8",
	Short: "Kubernetes release optimizer",
	Long: `
Iter8 is the Kubernetes release optimizer built for DevOps, MLOps, SRE and data science teams. Iter8 makes it easy to ensure that Kubernetes apps and ML models perform well and maximize business value.
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

// initialize Iter8 CLI root command and add all subcommands
func init() {
	// disable completion command for now
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "info", "trace, debug, info, warning, error, fatal, panic")
	rootCmd.SilenceErrors = true // will get printed in Execute() (by cobra.CheckErr())

	// add docs
	rootCmd.AddCommand(newDocsCmd())

	// add k
	rootCmd.AddCommand(kcmd)

	// add version
	rootCmd.AddCommand(newVersionCmd())

	// add controllers
	rootCmd.AddCommand(newControllersCmd(nil, kubeClient))

}
