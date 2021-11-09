package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/iter8-tools/iter8/base"
)

// formatText provides a text description of the experiment
func formatText(e *experiment) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	e.printState(w)
	if e.containsInsight(base.InsightTypeSLO) {
		if e.printableSLOs() {
			e.printSLOs(w)
		} else {
			e.printNoSLOs(w)
		}
	}
	if e.containsInsight(base.InsightTypeMetrics) {
		if e.printableMetrics() {
			e.printMetrics(w)
		} else {
			e.printNoMetrics(w)
		}
	}
	return b.String()
}

// number of completed tasks in the experiment
func (e *experiment) numCompletedTasksString() string {
	if e == nil || e.Result == nil {
		return "unknown"
	} else {
		return fmt.Sprint(e.Result.NumCompletedTasks)
	}
}

// print the current state of the experiment
func (e *experiment) printState(w *tabwriter.Writer) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "Experiment summary\t")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "Experiment completed \t"+strconv.FormatBool(e.completed()))
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "Experiment failed \t"+strconv.FormatBool(!e.noFailure()))
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "Number of completed tasks \t"+e.numCompletedTasksString())
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "")
	w.Flush()
}

// containsInsight checks if the experiment contains insight
func (e *experiment) containsInsight(in base.InsightType) bool {
	if e != nil {
		if e.Result != nil {
			if e.Result.Insights != nil {
				if e.Result.Insights.InsightTypes != nil {
					for _, v := range e.Result.Insights.InsightTypes {
						if v == in {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// are SLOs in a printable condition in this experiment
func (e *experiment) printableSLOs() bool {
	if e != nil {
		if e.Result != nil {
			if e.Result.Insights != nil {
				if len(e.Result.Insights.SLOStrs) > 0 {
					if *e.Result.Insights.NumAppVersions > 0 {
						if len(e.Result.Insights.SLOsSatisfied) == len(e.Result.Insights.SLOStrs) {
							if e.Result.Insights.SLOsSatisfied[0] != nil {
								if len(e.Result.Insights.SLOsSatisfied[0]) == *e.Result.Insights.NumAppVersions {
									return true
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

// print SLOs
func (e *experiment) printSLOs(w *tabwriter.Writer) {
	in := e.Result.Insights
	fmt.Fprint(w, "\n\n\n")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprint(w, "SLOs")
	if *in.NumAppVersions > 1 {
		for i := 0; i < *in.NumAppVersions; i++ {
			fmt.Fprintf(w, "\t version %v", i)
		}
	} else {
		fmt.Fprintf(w, "\t")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--------------------------\t-----")

	for i := 0; i < len(in.SLOStrs); i++ {
		fmt.Fprint(w, in.SLOStrs[i])
		for j := 0; j < *in.NumAppVersions; j++ {
			fmt.Fprintf(w, "\t%v", in.SLOsSatisfied[i][j])
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "--------------------------\t-----")
	}

	w.Flush()
}

// print no SLOs
func (e *experiment) printNoSLOs(w *tabwriter.Writer) {
	fmt.Fprint(w, "\n\n\n")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprint(w, "SLOs\tunavailable")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--------------------------\t-----")

	w.Flush()
}

// are metrics in a printable condition in this experiment
func (e *experiment) printableMetrics() bool {
	if e != nil {
		if e.Result != nil {
			if e.Result.Insights != nil {
				if e.Result.Insights.MetricsInfo != nil && len(e.Result.Insights.MetricsInfo) > 0 {
					if *e.Result.Insights.NumAppVersions > 0 {
						if len(e.Result.Insights.MetricValues) > 0 {
							if e.Result.Insights.MetricValues[0] != nil {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

// print metrics collected
func (e *experiment) printMetrics(w *tabwriter.Writer) {
	in := e.Result.Insights
	fmt.Fprint(w, "\n\n\n")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprint(w, "Metrics")
	if *in.NumAppVersions > 1 {
		for i := 0; i < *in.NumAppVersions; i++ {
			fmt.Fprintf(w, "\t version %v", i)
		}
	} else {
		fmt.Fprintf(w, "\t")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--------------------------\t-----")

	// sort metrics
	keys := []string{}
	for k := range in.MetricsInfo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < len(keys); i++ {
		fmt.Fprint(w, keys[i])
		for j := 0; j < *in.NumAppVersions; j++ {
			fmt.Fprintf(w, "\t%v", e.getMetricValueWithUnits(keys[i], j))
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "--------------------------\t-----")
	}
	w.Flush()
}

// print no metrics
func (e *experiment) printNoMetrics(w *tabwriter.Writer) {
	fmt.Fprint(w, "\n\n\n")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprint(w, "Metrics\tunavailable")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--------------------------\t-----")

	w.Flush()
}

// get value of the metric with units
func (e *experiment) getMetricValueWithUnits(m string, j int) string {
	vals := e.Result.Insights.MetricValues[j][m]
	if len(vals) == 0 {
		return "unavailable"
	}
	// get the latest observed value for this metric/version pair
	floatVal := vals[len(vals)-1]
	val := fmt.Sprint(floatVal)
	// if the floatVal is not integral, take two decimal places
	if floatVal != float64(int(floatVal)) {
		val = fmt.Sprintf("%0.2f", floatVal)
	}
	// add units if available
	units := e.Result.Insights.MetricsInfo[m].Units
	if units != nil {
		val += " (" + *units + ")"
	}
	return val
}
