package assert

import (
	"flag"
	"fmt"
	"os"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var example = `
	# assert that the most recent experiment running in the Kubernetes context is complete
	iter8 assert --remote -c completed
`

func NewCmd() *cobra.Command {
	cmd := basecli.AssertCmd
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)

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

	o := newOptions(streams)

	cmd.RunE = func(c *cobra.Command, args []string) error {
		if err := o.complete(factory, c, args); err != nil {
			return err
		}
		if err := o.validate(c, args); err != nil {
			return err
		}
		if err := o.run(c, args); err != nil {
			return err
		}
		return nil
	}

	cmd.Flags().StringVarP(&o.experiment, "experiment", "e", "", "remote experiment; if not specified, the most recent experiment is used")
	cmd.Flags().BoolVarP(&o.remote, "remote", "r", false, "test assertions on remotely executed experiment")

	return cmd
}
