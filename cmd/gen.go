package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// values are user specified values used during gen
	values []string
)

// GenCmd represents the gen command
var GenCmd = &cobra.Command{
	Use:   "gen",
	Short: "render templates with values",
	Long: `
	Render templates with values`,
}

func init() {
	RootCmd.AddCommand(GenCmd)
	GenCmd.PersistentFlags().StringSliceVarP(&values, "set", "s", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
}
