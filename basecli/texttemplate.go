package basecli

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	_ "embed"

	"github.com/iter8-tools/iter8/base/log"
)

// reportText is the text report template
//go:embed textreport.tpl
var reportText string

// PrintSLOsText returns SLOs in text report format
func (e *Experiment) PrintSLOsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	e.printSLOsText(w)
	return b.String()
}

func (e *Experiment) getSLOStrText(i int) (string, error) {
	in := e.Result.Insights
	slo := in.SLOs[i]
	// get metric with units and description
	str, err := e.MetricWithUnits(slo.Metric)
	if err != nil {
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
func (e *Experiment) printSLOsText(w *tabwriter.Writer) {
	in := e.Result.Insights
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
		str, err := e.getSLOStrText(i)
		if err == nil {
			fmt.Fprint(w, str)
			for j := 0; j < in.NumVersions; j++ {
				fmt.Fprintf(w, "\t%v", in.SLOsSatisfied[i][j])
				fmt.Fprintln(w)
			}
		}
	}

	w.Flush()
}

// PrintMetricsText returns metrics in text report format
func (e *Experiment) PrintMetricsText() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	e.printMetricsText(w)
	return b.String()
}

// printMetricsText prints metrics into tab writer
func (e *Experiment) printMetricsText(w *tabwriter.Writer) {
	in := e.Result.Insights
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
	keys := e.SortedScalarMetrics()

	for _, mn := range keys {
		mwu, err := e.MetricWithUnits(mn)
		if err == nil {
			// add metric name with units
			fmt.Fprint(w, mwu)
			// add value
			for j := 0; j < in.NumVersions; j++ {
				fmt.Fprintf(w, "\t%v", e.ScalarMetricValueStr(j, mn))
			}
			fmt.Fprintln(w)
		} else {
			log.Logger.Error(err)
		}
	}
	w.Flush()
}
