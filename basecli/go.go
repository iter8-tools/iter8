package basecli

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
	// Path to go template file
	goTemplateFilePath = "go.tpl"
)

func ParseValues(values []string, v chartutil.Values) error {
	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, v); err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("failed parsing --set data")
			return errors.Wrap(err, "failed parsing --set data")
		}
	}
	return nil
}

// goCmd represents the go command
var goCmd = &cobra.Command{
	Use:   "go",
	Short: "render a custom go template in the file go.tpl with values",
	Long: `
	Render a custom go template in the file go.tpl with values`,
	Example: `
	# render go template in go.tpl with values
	iter8 gen go --set key=val
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v := chartutil.Values{}
		err := ParseValues(GenOptions.Values, v)
		if err != nil {
			return err
		}
		// generate formatted output
		b, err := RenderGoTpl(v, goTemplateFilePath)
		if err != nil {
			return err
		}
		fmt.Println(b.String())
		return nil
	},
}

// RenderGoTpl creates output from go.tpl
func RenderGoTpl(v chartutil.Values, filePath string) (*bytes.Buffer, error) {
	var tmpl *template.Template
	var err error

	// read in the template file
	tplBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read template file")
		return nil, err
	}

	// add toYAML and other sprig template functions
	// they are all allowed to be used within the custom template
	// ensure it is a valid template
	tmpl, err = template.New("tpl").Funcs(template.FuncMap{
		"toYAML": toYAML,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(string(tplBytes))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
		return nil, err
	}

	// execute template
	var b bytes.Buffer
	err = tmpl.Execute(&b, v)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
		return nil, err
	}

	// print output
	return &b, nil
}

func init() {
	genCmd.AddCommand(goCmd)
}
