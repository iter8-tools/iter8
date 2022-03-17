package report

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"
	textT "text/template"

	_ "embed"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
)

type TextReporter struct {
	*Reporter
}

// reportText is the text report template
//go:embed textreport.tpl
var reportText string

func (tr *TextReporter) Gen(out io.Writer) error {
	// create text template
	ttpl, err := textT.New("report").Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(reportText)
	if err != nil {
		e := errors.New("unable to parse text template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	var b bytes.Buffer
	if err = ttpl.Execute(&b, tr); err != nil {
		e := errors.New("unable to execute template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// print output
	fmt.Fprintln(out, b.String())
	return nil
}

// PrintSLOsText returns SLOs in text report format
func (r *TextReporter) PrintSLOsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	r.printSLOsText(w)
	return b.String()
}

func (r *TextReporter) getSLOStrText(i int) (string, error) {
	in := r.Result.Insights
	slo := in.SLOs[i]
	// get metric with units and description
	str, err := r.MetricWithUnits(slo.Metric)
	if err != nil {
		log.Logger.Error("unable to get slo metric with units")
		return "", err
	}
	// add lower limit if needed
	if slo.LowerLimit != nil {
		str = fmt.Sprintf("%v <= %v", *slo.LowerLimit, str)
	}
	// add upper limit if needed
	if slo.UpperLimit != nil {
		str = fmt.Sprintf("%v <= %v", str, *slo.UpperLimit)
	}
	return str, nil
}

// printSLOsText prints SLOs into tab writer
func (r *TextReporter) printSLOsText(w *tabwriter.Writer) {
	in := r.Result.Insights
	fmt.Fprint(w, "SLO Conditions")
	if in.NumVersions > 1 {
		for i := 0; i < in.NumVersions; i++ {
			fmt.Fprintf(w, "\t version %v", i)
		}
	} else {
		fmt.Fprintf(w, "\tSatisfied")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--------------\t---------")

	for i := 0; i < len(in.SLOs); i++ {
		str, err := r.getSLOStrText(i)
		if err == nil {
			fmt.Fprint(w, str)
			for j := 0; j < in.NumVersions; j++ {
				fmt.Fprintf(w, "\t%v", in.SLOsSatisfied[i][j])
				fmt.Fprintln(w)
			}
		} else {
			log.Logger.Error("unable to extract SLO text")
		}
	}

	w.Flush()
}

// PrintMetricsText returns metrics in text report format
func (r *TextReporter) PrintMetricsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	r.printMetricsText(w)
	return b.String()
}

// printMetricsText prints metrics into tab writer
func (r *TextReporter) printMetricsText(w *tabwriter.Writer) {
	in := r.Result.Insights
	fmt.Fprint(w, "Metric")
	if in.NumVersions > 1 {
		for i := 0; i < in.NumVersions; i++ {
			fmt.Fprintf(w, "\tversion %v", i)
		}
	} else {
		fmt.Fprintf(w, "\tvalue")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "-------\t-----")

	// keys contain normalized scalar metric names in sorted order
	keys := r.SortedScalarAndSLOMetrics()

	for _, mn := range keys {
		mwu, err := r.MetricWithUnits(mn)
		if err == nil {
			// add metric name with units
			fmt.Fprint(w, mwu)
			// add value
			for j := 0; j < in.NumVersions; j++ {
				fmt.Fprintf(w, "\t%v", r.ScalarMetricValueStr(j, mn))
			}
			fmt.Fprintln(w)
		} else {
			log.Logger.Error(err)
		}
	}
	w.Flush()
}
