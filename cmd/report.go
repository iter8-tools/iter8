package cmd

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base"
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
	ttpl, err := template.New(TextOutputFormatKey).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(reportText)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse text template")
		os.Exit(1)
	}
	// register text template
	RegisterTextTemplate(TextOutputFormatKey, ttpl)

	// create HTML template
	htpl, err := htemplate.New(HTMLOutputFormatKey).Option("missingkey=error").Funcs(sprig.FuncMap()).Funcs(htemplate.FuncMap{
		"renderSLOSatisfiedHTML":      renderSLOSatisfiedHTML,
		"renderSLOSatisfiedCellClass": renderSLOSatisfiedCellClass,
	}).Parse(reportHTML)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse html template")
		os.Exit(1)
	}
	// register HTML template
	RegisterHTMLTemplate(HTMLOutputFormatKey, htpl)

	reportCmd = NewReportCmd()
	RootCmd.AddCommand(reportCmd)
}

/* Following functions/methods are common to both text and html templates */

// SortedScalarAndSLOMetrics extracts scalar and SLO metric names from experiment in sorted order
func (e *Experiment) SortedScalarAndSLOMetrics() []string {
	keys := []string{}
	for k, mm := range e.Result.Insights.MetricsInfo {
		if mm.Type == base.CounterMetricType || mm.Type == base.GaugeMetricType {
			keys = append(keys, k)
		}
	}
	// also add SLO metric names
	for _, v := range e.Result.Insights.SLOs {
		nm, err := base.NormalizeMetricName(v.Metric)
		if err == nil {
			keys = append(keys, nm)
		}
	}
	// remove duplicates
	tmp := base.Uniq(keys)
	uniqKeys := []string{}
	for _, val := range tmp {
		uniqKeys = append(uniqKeys, val.(string))
	}

	sort.Strings(uniqKeys)
	return uniqKeys
}

// ScalarMetricValueStr extracts metric value string for given version and scalar metric name
func (e *Experiment) ScalarMetricValueStr(j int, mn string) string {
	val := e.Result.Insights.ScalarMetricValue(j, mn)
	if val != nil {
		return fmt.Sprintf("%0.2f", *val)
	} else {
		return "unavailable"
	}
}
