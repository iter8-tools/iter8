package cmd

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	assert "github.com/iter8-tools/iter8/k8s/pkg/cmd/assert"
	gen "github.com/iter8-tools/iter8/k8s/pkg/cmd/gen"
	hub "github.com/iter8-tools/iter8/k8s/pkg/cmd/hub"
	report "github.com/iter8-tools/iter8/k8s/pkg/cmd/report"
	run "github.com/iter8-tools/iter8/k8s/pkg/cmd/run"

	deleter "github.com/iter8-tools/iter8/k8s/pkg/cmd/deleter"
	get "github.com/iter8-tools/iter8/k8s/pkg/cmd/get"
)

func NewCmdIter8Command() *cobra.Command {
	root := &cobra.Command{
		Use:   "iter8",
		Short: "Manage an experiment",
		Long: templates.LongDesc(`
      Run and inspect an Iter8 experiment.

      Find more information at:
            https://iter8.tools/`),
	}

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
	// root.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// root.AddCommand(cmdconfig.NewCmdConfig(f, clientcmd.NewDefaultPathOptions(), streams))
	root.AddCommand(options.NewCmdOptions(streams.Out))

	// //enable plugin functionality: all `os.Args[0]-<binary>` in the $PATH will be available for plugin
	// plugin.ValidPluginFilenamePrefixes = []string{os.Args[0]}
	// root.AddCommand(plugin.NewCmdPlugin(streams))

	groups := templates.CommandGroups{
		{
			Message: "Current Commands:",
			Commands: []*cobra.Command{
				hub.NewCmd(),
				run.NewCmd(factory, streams),
				gen.NewCmd(),
				get.NewCmd(factory, streams),
				assert.NewCmd(factory, streams),
				report.NewCmd(factory, streams),
			},
		},
		{
			Message: "Deprecated commands:",
			Commands: []*cobra.Command{
				deleter.NewCmd(factory, streams),
			},
		},
	}
	groups.Add(root)

	filters := []string{}

	templates.ActsAsRootCommand(root, filters, groups...)

	return root
}
