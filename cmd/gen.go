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

var (
	// Path to template file
	// this variable is intended to be modified in tests, and nowhere else
	templateFilePath = "iter8.tpl"

	// Output format
	outputFormat string = "text"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "format experiment spec and its result using a go template",
	Example: `
	# generate text output of the experiment using the built-in Go template
	iter8 gen -o text
	# iter8 gen does the same thing as above

	# generate output from experiment using a custom Go template specified in the iter8.tpl file
	iter8 gen -o custom`,
	Run: func(cmd *cobra.Command, args []string) {
		var tmpl *template.Template
		var err error

		switch strings.ToLower(outputFormat) {
		case "text":
			tmpl, err = builtInTemplates["text"].Parse("{{ formatText . }}")
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to parse text template")
				os.Exit(1)
			}

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
}
