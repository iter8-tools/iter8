package cmd

import (
	"flag"
	"os"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewCmdKCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "k",
		Short: "Work with experiments running in a Kubernetes cluster",
		Example: `
To run an experiment defined in 'experiment.yaml':
iter8 gen k8s | kubectl apply -f -

To delete an experiment with identifier $EXPERIMENT_ID:
iter8 gen k8s --set id=$EXPERIMENT_ID | kubectl delete -f -`,
		// There is no action associated with this command
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	// Add the default kubectl options as persistent flags
	flags := root.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)

	matchVersionKubeConfigFlags.AddFlags(flags)
	flags.AddGoFlagSet(flag.CommandLine)

	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// // From this point and forward we get warnings on flags that contain "_" separators
	// cmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// This adds the "config" subcommand that allows changes to kubeconfig files
	// cmd.AddCommand(cmdconfig.NewCmdConfig(clientcmd.NewDefaultPathOptions(), streams))

	// Modify the help for this command to hide the k8s specific flags by default
	// Provide 'options' command to display them
	help := root.HelpFunc()
	root.SetHelpFunc(func(command *cobra.Command, strings []string) {
		// Hide flags for this command
		command.PersistentFlags().VisitAll(func(f *pflag.Flag) { command.PersistentFlags().MarkHidden(f.Name) })
		// Call the cached help function
		help(command, strings)
	})
	root.AddCommand(options.NewCmdOptions(streams.Out))

	// Include the valid subcommands for 'k':
	root.AddCommand(NewRunCmd(factory, streams))
	root.AddCommand(NewGetCmd(factory, streams))
	root.AddCommand(NewAssertCmd(factory, streams))
	root.AddCommand(NewReportCmd(factory, streams))

	// extend base gen command with the k8s command
	basecli.GenCmd.AddCommand(NewGetK8sCmd())

	return root
}
