package cmd

import (
	"errors"
	"io"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// klaunchDesc is the description of the k launch cmd
const klaunchDesc = `
Launch an experiment inside a Kubernetes cluster. 

	iter8 k launch --set "tasks={http}" --set http.url=https://httpbin.org/get \
	--set runner=job

Use the dry option to simulate a Kubernetes experiment. This creates the manifest.yaml file, but does not run the experiment, and does not deploy any experiment resource objects in the cluster.

	iter8 k launch \
	--set http.url=https://httpbin.org/get \
	--set runner=job \
	--dry

The launch command creates the 'charts' subdirectory under the current working directory, downloads the Iter8 experiment chart, and places it under 'charts'. This behavior can be controlled using various launch flags.

This command supports setting values using the same mechanisms as in Helm. Please see  https://helm.sh/docs/chart_template_guide/values_files/ for more detailed descriptions. In particular, this command supports the --set, --set-file, --set-string, and -f (--values) options all of which have the same behavior as in Helm.
`

// newKLaunchCmd creates the Kubernetes launch command
func newKLaunchCmd(kd *driver.KubeDriver, out io.Writer) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:          "launch",
		Short:        "Launch an experiment inside a Kubernetes cluster",
		Long:         klaunchDesc,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return chartNameIsRequired(actor, cmd.Flags())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.KubeRun()
		},
	}
	// flags specific to k launch
	addExperimentGroupFlag(cmd, &actor.Group)
	addDryRunForKFlag(cmd, &actor.DryRun)
	actor.EnvSettings = settings

	// flags shared with launch
	// addChartPathOptionsFlags(cmd, &actor.ChartPathOptions)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	addLocalChartFlag(cmd, &actor.LocalChart)

	return cmd
}

// chartNameIsRequired makes chartName required if localChart is set
func chartNameIsRequired(lOpts *ia.LaunchOpts, flags *pflag.FlagSet) error {
	if flags.Changed("localChart") && !flags.Changed("chartName") {
		return errors.New("localChart specified; 'chartName' is required")
	}
	return nil
}

// addDryRunForKFlag adds dry run flag to the k launch command
func addDryRunForKFlag(cmd *cobra.Command, dryRunPtr *bool) {
	cmd.Flags().BoolVar(dryRunPtr, "dry", false, "simulate an experiment launch; outputs manifest.yaml file")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

// addChartNameFlag to the command
func addChartNameFlag(cmd *cobra.Command, chartNamePtr *string) {
	cmd.Flags().StringVarP(chartNamePtr, "chartName", "c", ia.DefaultChartName, "name of the experiment chart")
}

// addLocalChartFlag adds the localChart flag to the launch command
func addLocalChartFlag(cmd *cobra.Command, localChartPtr *bool) {
	cmd.Flags().BoolVar(localChartPtr, "localChart", false, "use local chart identified by --chartName")
	cmd.Flags().Lookup("localChart").NoOptDefVal = "true"
}

// addChartPathOptionsFlags adds flags related to Helm chart repository
// copied from
// https://github.com/helm/helm/blob/ce66412a723e4d89555dc67217607c6579ffcb21/cmd/helm/flags.go
// func addChartPathOptionsFlags(cmd *cobra.Command, c *action.ChartPathOptions) {
// cmd.Flags().StringVar(&c.Version, "version", "", "specify a version constraint for the chart version to use. This constraint can be a specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is not specified, the latest version is used")
// cmd.Flags().BoolVar(&c.Verify, "verify", false, "verify the package before using it")
// cmd.Flags().StringVar(&c.Keyring, "keyring", defaultKeyring(), "location of public keys used for verification")
// cmd.Flags().StringVar(&c.RepoURL, "repo", "https://iter8-tools.github.io/iter8", "chart repository url where to locate the requested chart")
// cmd.Flags().StringVar(&c.Username, "username", "", "chart repository username where to locate the requested chart")
// cmd.Flags().StringVar(&c.Password, "password", "", "chart repository password where to locate the requested chart")
// cmd.Flags().StringVar(&c.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
// cmd.Flags().StringVar(&c.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file")
// cmd.Flags().BoolVar(&c.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "skip tls certificate checks for the chart download")
// cmd.Flags().StringVar(&c.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
// cmd.Flags().BoolVar(&c.PassCredentialsAll, "pass-credentials", false, "pass credentials to all domains")
// }

// // defaultKeyring returns the expanded path to the default keyring.
// // copied from
// // https://github.com/helm/helm/blob/ce66412a723e4d89555dc67217607c6579ffcb21/cmd/helm/dependency_build.go
// func defaultKeyring() string {
// 	if v, ok := os.LookupEnv("GNUPGHOME"); ok {
// 		return filepath.Join(v, "pubring.gpg")
// 	}
// 	return filepath.Join(homedir.HomeDir(), ".gnupg", "pubring.gpg")
// }
