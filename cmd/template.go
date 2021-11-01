package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/core/log"
	"github.com/spf13/cobra"
)

var (
	// Path to template file
	// this variable is intended to be modified in tests, and nowhere else
	filePath = "iter8.tpl"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:    "template",
	Short:  "generate textual output of the experiment using a Go template",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// read in the template file
		tplBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to read template file")
			os.Exit(1)
		}

		// ensure it is a valid template
		tmpl, err := template.New("tpl").Funcs(template.FuncMap{
			"toYAML": toYAML,
		}).Option("missingkey=error").Funcs(sprig.FuncMap()).Parse(string(tplBytes))
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
			os.Exit(1)
		}

		// build experiment

		log.Logger.Trace("build started")
		exp, err := Build(false)
		log.Logger.Trace("build finished")
		if err != nil {
			log.Logger.Error("experiment build failed")
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
		fmt.Println(b.String())
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
