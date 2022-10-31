package cmd

import (
	"errors"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/driver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	// "helm.sh/helm/v3/pkg/action"
	// "k8s.io/client-go/util/homedir"
)

// launchDesc is the description of the launch command
const launchDesc = `
Launch an experiment in the local environment. 

	iter8 launch --set "tasks={http}" \
	--set http.url=https://httpbin.org/get

Use the dry option to simulate an experiment. This creates the experiment.yaml file but does not run the experiment.

	iter8 launch \
	--set http.url=https://httpbin.org/get \
	--dry

The launch command creates the 'charts' subdirectory under the current working directory, downloads the Iter8 experiment chart, and places it under 'charts'. This behavior can be controlled using various launch flags.

This command supports setting values using the same mechanisms as in Helm. Please see  https://helm.sh/docs/chart_template_guide/values_files/ for more detailed descriptions. In particular, this command supports the --set, --set-file, --set-string, and -f (--values) options all of which have the same behavior as in Helm.
`

// newLaunchCmd creates the launch command
func newLaunchCmd(kd *driver.KubeDriver) *cobra.Command {
	actor := ia.NewLaunchOpts(kd)

	cmd := &cobra.Command{
		Use:          "launch",
		Short:        "Launch an experiment in the local environment",
		Long:         launchDesc,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return chartNameIsRequired(actor, cmd.Flags())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return actor.LocalRun()
		},
	}
	addDryRunFlag(cmd, &actor.DryRun)
	// addChartPathOptionsFlags(cmd, &actor.ChartPathOptions)
	addChartNameFlag(cmd, &actor.ChartName)
	addValueFlags(cmd.Flags(), &actor.Options)
	addRunDirFlag(cmd, &actor.RunDir)
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

// addDryRunFlag adds dry run flag to the launch command
func addDryRunFlag(cmd *cobra.Command, dryRunPtr *bool) {
	cmd.Flags().BoolVar(dryRunPtr, "dry", false, "simulate an experiment launch; outputs experiment.yaml file")
	cmd.Flags().Lookup("dry").NoOptDefVal = "true"
}

// addLocalChartFlag adds the localChart flag to the launch command
func addLocalChartFlag(cmd *cobra.Command, localChartPtr *bool) {
	cmd.Flags().BoolVar(localChartPtr, "localChart", false, "use local charts")
	cmd.Flags().Lookup("localChart").NoOptDefVal = "true"
}

// addChartPathOptionsFlags adds flags related to Helm chart repository
// copied from
// https://github.com/helm/helm/blob/ce66412a723e4d89555dc67217607c6579ffcb21/cmd/helm/flags.go
// func addChartPathOptionsFlags(cmd *cobra.Command, c *action.ChartPathOptions) {
// cmd.Flags().StringVar(&c.Version, "version", "", "specify a version constraint for the chart version to use. This constraint can be a specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is not specified, the latest version is used")
// cmd.Flags().BoolVar(&c.Verify, "verify", false, "verify the package before using it")
// cmd.Flags().StringVar(&c.Keyring, "keyring", defaultKeyring(), "location of public keys used for verification")
// cmd.Flags().StringVar(&c.RepoURL, "repo", "https://iter8-tools.github.io/hub", "chart repository url where to locate the requested chart")
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
