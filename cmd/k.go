package cmd

// import (
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/pflag"
// )

// var k8sExperimentOptions = newK8sExperimentOptions()

// var kCmd = &cobra.Command{
// 	Use:   "k",
// 	Short: "Work with experiments running in Kubernetes",
// 	Example: `
// To generate a Kubernetes manifest for an experiment in 'experiment.yaml',
// and run it in Kubernetes, do:

// 	iter8 gen k8s | kubectl apply -f -
// `,
// 	// There is no action associated with this command
// 	// Run: func(cmd *cobra.Command, args []string) { },
// }

// var sharedFlags = pflag.NewFlagSet("kflags", pflag.ExitOnError)

// // getIdFlag returns the id flag.
// // This function enables reuse of this flag across subcommands.
// func getIdFlag() *pflag.Flag {
// 	if sharedFlags.Lookup("id") == nil {
// 		sharedFlags.StringVarP(k8sExperimentOptions.id, "id", "i", "", "experiment identifier; if not specified, the most recent experiment is used")
// 	}
// 	return sharedFlags.Lookup("id")
// }

// // getAppFlag returns the app flag.
// // This function enables reuse of this flag across subcommands.
// func getAppFlag() *pflag.Flag {
// 	if sharedFlags.Lookup("app") == nil {
// 		sharedFlags.StringVarP(k8sExperimentOptions.app, "app", "a", "", "app label; this flag is ignored if --id flag is specified")
// 	}
// 	return sharedFlags.Lookup("app")
// }

// func init() {
// 	RootCmd.AddCommand(kCmd)
// 	flags := kCmd.PersistentFlags()
// 	k8sExperimentOptions.ConfigFlags.AddFlags(flags)
// }
