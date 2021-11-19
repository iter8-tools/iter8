package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
)

const (
	// Path to template file
	templateFilePath = "iter8.tpl"
)

var (
	// Output format
	outputFormat string = "text"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate formatted output from experiment spec and result",
	Long:  "Generate formatted output from experiment spec and result",
	Example: `
	# download the load-test experiment
	iter8 hub -e load-test

	cd load-test

	# run it
	iter8 run

	# generate formatted text output
	iter8 gen
`,
	Run: func(cmd *cobra.Command, args []string) {
		var tmpl *template.Template
		var err error

		switch strings.ToLower(outputFormat) {
		case "text":
			tmpl = builtInTemplates["text"]

		case "custom":
			// read in the template file
			tplBytes, err := ioutil.ReadFile(templateFilePath)
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to read template file")
				os.Exit(1)
			}

			// ensure it is a valid template
			tmpl, err = template.New("tpl").Funcs(template.FuncMap{
				"toYAML": toYAML,
			}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(string(tplBytes))
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
				os.Exit(1)
			}

		default:
			log.Logger.Error("invalid output format; valid formats are: text | custom")
			os.Exit(1)
		}

		// build experiment
		log.Logger.Trace("build started")
		exp, err := build(true)
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
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().StringVarP(&outputFormat, "outputFormat", "o", "text", "text | custom")
	genCmd.Flags().MarkHidden("outputFormat")
}
