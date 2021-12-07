package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	// TextOutputFormat is the output format used to create text output
	TextOutputFormatKey = "text"

	// HTMLOutputFormat is the output format used to create html output
	HTMLOutputFormatKey = "html"
)

// executable
type executable interface {
	Execute(w io.Writer, data interface{}) error
}

type ReportOptionsType struct {
	// Output format variable holds the output format to be used by gen
	OutputFormat string
}

var ReportOptions = ReportOptionsType{
	OutputFormat: TextOutputFormatKey,
}

func NewReportCmd() *cobra.Command {
	// ReportCmd represents the report command
	var ReportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate report from experiment result",
		Long:  `Generate report from experiment result`,
		Example: `
# generate text report
iter8 report

# generate html report
iter8 report -o html
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Logger.Trace("build started")
			// build experiment
			// replace FileExpIO with ClusterExpIO to build from cluster
			fio := &FileExpIO{}
			exp, err := Build(true, fio)
			log.Logger.Trace("build finished")
			if err != nil {
				return err
			}

			// generate formatted output from experiment
			err = exp.Report(ReportOptions.OutputFormat)
			if err != nil {
				return err
			}
			return nil
		},
	}

	ReportCmd.Flags().StringVarP(&ReportOptions.OutputFormat, "outputFormat", "o", "text", "text | html")

	// create text template
	tmpl, err := template.New(TextOutputFormatKey).Funcs(template.FuncMap{
		"formatText": formatText,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse("{{ formatText . }}")
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse text template")
		os.Exit(1)
	}
	// register text template
	RegisterTextTemplate(TextOutputFormatKey, tmpl)

	// create HTML template (for now, this will still use the text templating functionality)
	htmpl, err := template.New(TextOutputFormatKey).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(formatHTML)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse html template")
		os.Exit(1)
	}
	// register HTML template
	RegisterTextTemplate(HTMLOutputFormatKey, htmpl)

	return ReportCmd
}

// Report creates a report from experiment as per outputFormat
func (exp *Experiment) Report(outputFormat string) error {
	templateKey := strings.ToLower(outputFormat)

	tmpl, ok := builtInTemplates[templateKey]
	if !ok {
		log.Logger.Error("invalid output format; valid formats are: text | html")
		return errors.New("invalid output format; valid formats are: text | html")
	}

	// execute template
	return ExecTemplate(tmpl, exp)
}

// execute text or html template with experiment
func ExecTemplate(t executable, exp *Experiment) error {
	var b bytes.Buffer
	err := t.Execute(&b, exp)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
		return err
	}

	// print output
	fmt.Println(b.String())
	return nil
}

func init() {
	RootCmd.AddCommand(NewReportCmd())
}
