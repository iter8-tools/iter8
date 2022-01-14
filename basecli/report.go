package basecli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
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

// ReportOptionsType enables options for the report command
type ReportOptionsType struct {
	// OutputFormat holds the output format to be used by report
	OutputFormat string
}

// ReportOptions stores the options used by thee report command
var ReportOptions = ReportOptionsType{
	OutputFormat: TextOutputFormatKey,
}

var reportCmd *cobra.Command

// NewReportCmd creates a new report command
func NewReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "View report from experiment result",
		Long:  `View report from experiment result`,
		Example: `
	# view text report
	iter8 report
	
	# view html report
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
	cmd.Flags().StringVarP(&ReportOptions.OutputFormat, "outputFormat", "o", "text", "text | html")

	return cmd
}

// Report creates a report from experiment as per outputFormat
func (exp *Experiment) Report(outputFormat string) error {
	templateKey := strings.ToLower(outputFormat)

	tmpl, ok := builtInTemplates[templateKey]
	if !ok {
		e := fmt.Errorf("invalid output format; valid formats are: %v | %v", TextOutputFormatKey, HTMLOutputFormatKey)
		log.Logger.Error(e)
		return e
	}

	// execute template
	return ExecTemplate(tmpl, exp)
}

// ExecTemplate executes text or html template using experiment as the data
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
	htmpl, err := template.New(TextOutputFormatKey).Funcs(template.FuncMap{
		"styleSection": styleSection,
		"headSection":  headSection,
		"dependencies": dependencies,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(formatHTML)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse html template")
		os.Exit(1)
	}

	// register HTML template
	RegisterTextTemplate(HTMLOutputFormatKey, htmpl)

	reportCmd = NewReportCmd()

	RootCmd.AddCommand(reportCmd)
}
