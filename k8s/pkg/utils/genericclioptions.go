package utils

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func AddGenericCliOptions(cmd *cobra.Command) (cmdutil.Factory, genericclioptions.IOStreams) {
	// Add the default kubectl options as persistent flags
	flags := cmd.PersistentFlags()
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

	// Add the "options" subcommand to display available options
	cmd.AddCommand(options.NewCmdOptions(streams.Out))

	return factory, streams
}
