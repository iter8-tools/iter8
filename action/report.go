package action

import (
	"bytes"
	"errors"
	"fmt"
	htmlT "html/template"
	"sort"
	"strings"
	textT "text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/action"
)

const (
	// TextOutputFormat is the output format used to create text output
	TextOutputFormatKey = "text"

	// HTMLOutputFormat is the output format used to create html output
	HTMLOutputFormatKey = "html"
)

type Reporter base.Experiment

type ReportOpts struct {
	// OutputFormat holds the output format to be used by report
	OutputFormat string
	// applicable only for local experiments
	RunOpts
	// applicable only for kubernetes experiments
	driver.ExperimentResource
}

func NewReportOpts(cfg *action.Configuration) *ReportOpts {
	return &ReportOpts{}
}

func (report *ReportOpts) LocalRun() error {
	return report.Run(&driver.FileDriver{
		RunDir: report.RunDir,
	})
}

func (report *ReportOpts) Run(eio base.Driver) error {
	if e, err := base.BuildExperiment(true, eio); err != nil {
		return err
	} else {
		reporter := Reporter(*e)
		switch strings.ToLower(report.OutputFormat) {
		case TextOutputFormatKey:
			return reporter.genText()
		case HTMLOutputFormatKey:
			return reporter.genHTML()
		default:
			e := fmt.Errorf("unsupported report format %v", report.OutputFormat)
			log.Logger.Error(e)
			return e
		}
	}
}

func (reporter *Reporter) genText() error {
	// reportText is the text report template
	//go:embed textreport.tpl
	var reportText string

	// create text template
	ttpl, err := textT.New(TextOutputFormatKey).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(reportText)
	if err != nil {
		e := errors.New("unable to parse text template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	var b bytes.Buffer
	if err = ttpl.Execute(&b, reporter); err != nil {
		e := errors.New("unable to execute template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// print output
	fmt.Println(b.String())
	return nil
}

func (reporter *Reporter) genHTML() error {
	// reportHTML is the HTML report template
	//go:embed htmlreport.tpl
	var reportHTML string

	// create HTML template
	htpl, err := htmlT.New(TextOutputFormatKey).Option("missingkey=error").Funcs(sprig.FuncMap()).Parse(reportHTML)
	if err != nil {
		e := errors.New("unable to parse HTML template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	var b bytes.Buffer
	if err = htpl.Execute(&b, reporter); err != nil {
		e := errors.New("unable to execute template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// print output
	fmt.Println(b.String())
	return nil
}

/* Following functions/methods are common to both text and html templates */

// SortedScalarAndSLOMetrics extracts scalar and SLO metric names from experiment in sorted order
func (r *Reporter) SortedScalarAndSLOMetrics() []string {
	keys := []string{}
	for k, mm := range r.Result.Insights.MetricsInfo {
		if mm.Type == base.CounterMetricType || mm.Type == base.GaugeMetricType {
			keys = append(keys, k)
		}
	}
	// also add SLO metric names
	for _, v := range r.Result.Insights.SLOs {
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
func (r *Reporter) ScalarMetricValueStr(j int, mn string) string {
	val := r.Result.Insights.ScalarMetricValue(j, mn)
	if val != nil {
		return fmt.Sprintf("%0.2f", *val)
	} else {
		return "unavailable"
	}
}
