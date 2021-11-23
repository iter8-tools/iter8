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
	GenCmd.PersistentFlags().StringSliceVarP(&values, "set", "s", []string{}, "key=value; value can be accessed in templates used by gen {{ Values.key }}")
}
