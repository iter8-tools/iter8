package basecli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/cli/values"
)

// GenOptions are the options used by the gen subcommands.
// They store values that can be combined with templates for generating experiment.yaml files Kubernetes manifests.
var GenOptions = values.Options{}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Render templates with values",
	Long: `
Render templates with values`,
}

func addGenOptions(f *pflag.FlagSet) {
	// See: https://github.com/helm/helm/blob/663a896f4a815053445eec4153677ddc24a0a361/cmd/helm/flags.go#L42 which is the source of these flags
	f.StringSliceVarP(&GenOptions.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&GenOptions.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&GenOptions.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&GenOptions.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
}

func init() {
	RootCmd.AddCommand(genCmd)
	addGenOptions(genCmd.PersistentFlags())
}
