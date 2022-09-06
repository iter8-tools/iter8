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
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

// TextReporter supports generation of text reports from experiments.
type TextReporter struct {
	// Reporter is embedded and enables access to all reporter data and methods
	*Reporter
}

// reportText is the text report template
//go:embed textreport.tpl
var reportText string

// Gen writes the text report for a given experiment into the given writer
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

// PrintSLOsText returns SLOs section of the text report as a string
func (r *TextReporter) PrintSLOsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	r.printSLOsText(w)
	return b.String()
}

// getSLOStrText gets the text for an SLO
func (r *TextReporter) getSLOStrText(i int, upper bool) (string, error) {
	in := r.Result.Insights
	var slo base.SLO
	if upper {
		slo = in.SLOs.Upper[i]
	} else {
		slo = in.SLOs.Lower[i]
	}
	// get metric with units and description
	str, err := r.MetricWithUnits(slo.Metric)
	if err != nil {
		log.Logger.Error("unable to get slo metric with units")
		return "", err
	}
	// add upper limit
	if upper {
		str = fmt.Sprintf("%v <= %v", str, slo.Limit)
	} else {
		// add lower limit
		str = fmt.Sprintf("%v <= %v", slo.Limit, str)
	}
	return str, nil
}

// printSLOsText prints all SLOs into tab writer
func (r *TextReporter) printSLOsText(w *tabwriter.Writer) {
	in := r.Result.Insights
	fmt.Fprint(w, "SLO Conditions")
	if in.NumVersions > 1 {
		for i := 0; i < in.NumVersions; i++ {
			fmt.Fprintf(w, "\t version %v", i)
		}
	} else {
		fmt.Fprintf(w, "\t Satisfied")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--------------\t ---------")

	if in.SLOs != nil {
		for i := 0; i < len(in.SLOs.Upper); i++ {
			str, err := r.getSLOStrText(i, true)
			if err == nil {
				fmt.Fprint(w, str)
				for j := 0; j < in.NumVersions; j++ {
					fmt.Fprintf(w, "\t %v", in.SLOsSatisfied.Upper[i][j])
				}
				fmt.Fprintln(w)
			} else {
				log.Logger.Error("unable to extract SLO text")
			}
		}
		for i := 0; i < len(in.SLOs.Lower); i++ {
			str, err := r.getSLOStrText(i, false)
			if err == nil {
				fmt.Fprint(w, str)
				for j := 0; j < in.NumVersions; j++ {
					fmt.Fprintf(w, "\t %v", in.SLOsSatisfied.Lower[i][j])
				}
				fmt.Fprintln(w)
			} else {
				log.Logger.Error("unable to extract SLO text")
			}
		}
	}

	w.Flush()
}

// PrintMetricsText returns metrics section of the text report as a string
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
			fmt.Fprintf(w, "\t version %v", i)
		}
	} else {
		fmt.Fprintf(w, "\t value")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "-------\t -----")

	// keys contain normalized scalar metric names in sorted order
	keys := r.SortedScalarAndSLOMetrics()

	for _, mn := range keys {
		mwu, err := r.MetricWithUnits(mn)
		if err == nil {
			// add metric name with units
			fmt.Fprint(w, mwu)
			// add value
			for j := 0; j < in.NumVersions; j++ {
				fmt.Fprintf(w, "\t %v", r.ScalarMetricValueStr(j, mn))
			}
			fmt.Fprintln(w)
		} else {
			log.Logger.Error(err)
		}
	}
	w.Flush()
}
