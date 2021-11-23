package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	// Path to template file
	templateFilePath = "go.tpl"
)

func parseValues(values []string, v chartutil.Values) error {
	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, v); err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("failed parsing --set data")
			return errors.Wrap(err, "failed parsing --set data")
		}
	}
	return nil
}

// GoCmd represents the go command
var GoCmd = &cobra.Command{
	Use:   "go",
	Short: "render a custom go template with values",
	Long:  "Render a custom go template with values",
	Example: `
	# use go template in go.tpl
	# execute it using values that are set
	iter8 gen go --set a=b
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v := chartutil.Values{}
		err := parseValues(values, v)
		if err != nil {
			return err
		}
		// generate formatted output
		err = Go(v)
		if err != nil {
			return err
		}
		return nil
	},
}

// Go creates output from go.tpl
func Go(v chartutil.Values) error {
	var tmpl *template.Template
	var err error

	// read in the template file
	tplBytes, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read template file")
		return err
	}

	// add toYAML and other sprig template functions
	// they are all allowed to be used within the custom template
	// ensure it is a valid template
	tmpl, err = template.New("tpl").Funcs(template.FuncMap{
		"toYAML": toYAML,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(string(tplBytes))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
		return err
	}

	// execute template
	var b bytes.Buffer
	err = tmpl.Execute(&b, v)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
		return err
	}

	// print output
	fmt.Println(b.String())
	return nil
}

func init() {
	GenCmd.AddCommand(GoCmd)
}
