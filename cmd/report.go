package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "report results from experiment",
	Run: func(cmd *cobra.Command, args []string) {
		// build experiment
		log.Logger.Trace("build started")
		exp, err := build(true)
		log.Logger.Trace("build finished")
		if err != nil {
			log.Logger.Error("experiment build failed")
			os.Exit(1)
		}

		// execute template
		tmpl, err := builtInTemplates["txt"].Parse("{{ describeTxt . }}")
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to parse template")
			os.Exit(1)
		}

		// execute template
		var b bytes.Buffer
		err = tmpl.Execute(&b, exp)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
			os.Exit(1)
		}

		// print output
		fmt.Print(b.String())
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
