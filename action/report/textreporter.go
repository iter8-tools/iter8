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
//
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
func (tr *TextReporter) PrintSLOsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	tr.printSLOsText(w)
	return b.String()
}

// getSLOStrText gets the text for an SLO
func (tr *TextReporter) getSLOStrText(i int, upper bool) (string, error) {
	in := tr.Result.Insights
	var slo base.SLO
	if upper {
		slo = in.SLOs.Upper[i]
	} else {
		slo = in.SLOs.Lower[i]
	}
	// get metric with units and description
	str, err := tr.MetricWithUnits(slo.Metric)
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
func (tr *TextReporter) printSLOsText(w *tabwriter.Writer) {
	in := tr.Result.Insights
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

	if in.SLOs != nil {
		log.Logger.Debug("SLOs are not nil")
		log.Logger.Debug("found ", len(in.SLOs.Upper), " upper SLOs")
		for i := 0; i < len(in.SLOs.Upper); i++ {
			log.Logger.Debug("Upper SLO ", i)
			str, err := tr.getSLOStrText(i, true)
			if err == nil {
				fmt.Fprint(w, str)
				for j := 0; j < in.NumVersions; j++ {
					fmt.Fprintf(w, "\t%v", in.SLOsSatisfied.Upper[i][j])
					fmt.Fprintln(w)
				}
			} else {
				log.Logger.Error("unable to extract SLO text")
			}
		}

		log.Logger.Debug("found ", len(in.SLOs.Lower), " lower SLOs")
		for i := 0; i < len(in.SLOs.Lower); i++ {
			log.Logger.Debug("Lower SLO ", i)
			str, err := tr.getSLOStrText(i, false)
			if err == nil {
				fmt.Fprint(w, str)
				for j := 0; j < in.NumVersions; j++ {
					fmt.Fprintf(w, "\t%v", in.SLOsSatisfied.Lower[i][j])
					fmt.Fprintln(w)
				}
			} else {
				log.Logger.Error("unable to extract SLO text")
			}
		}
	}

	_ = w.Flush()
}

// PrintMetricsText returns metrics section of the text report as a string
func (tr *TextReporter) PrintMetricsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	tr.printMetricsText(w)
	return b.String()
}

// printMetricsText prints metrics into tab writer
func (tr *TextReporter) printMetricsText(w *tabwriter.Writer) {
	in := tr.Result.Insights
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
	keys := tr.SortedScalarAndSLOMetrics()

	for _, mn := range keys {
		mwu, err := tr.MetricWithUnits(mn)
		if err == nil {
			// add metric name with units
			fmt.Fprint(w, mwu)
			// add value
			for j := 0; j < in.NumVersions; j++ {
				fmt.Fprintf(w, "\t%v", tr.ScalarMetricValueStr(j, mn))
			}
			fmt.Fprintln(w)
		} else {
			log.Logger.Error(err)
		}
	}
	_ = w.Flush()
}
