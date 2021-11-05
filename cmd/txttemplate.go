package cmd

import (
	"bytes"
	"fmt"
	"strconv"
	"text/tabwriter"
)

// describeTxt provides a text description of the experiment
func describeTxt(e *experiment) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	e.printSummary(w)
	return b.String()
}

// print the testing pattern for the experiment
func (e *experiment) testingPatternString() string {
	if e == nil || e.Result == nil || e.Result.Analysis == nil || e.Result.Analysis.TestingPattern == nil {
		return "unknown"
	} else {
		return string(*e.Result.Analysis.TestingPattern)
	}
}

// number of completed tasks in the experiment
func (e *experiment) numCompletedTasksString() string {
	if e == nil || e.Result == nil {
		return "unknown"
	} else {
		return fmt.Sprintf("%v", e.Result.NumCompletedTasks)
	}
}

// winning version
func (e *experiment) winner() string {
	if e == nil || e.Result == nil || e.Result.Analysis == nil {
		return "unknown"
	} else if e.Result.Analysis.Winner == nil {
		return "not found"
	} else {
		return *e.Result.Analysis.Winner
	}
}

// number of app versions
func (e *experiment) numAppVersions() string {
	if e == nil || e.Result == nil || e.Result.NumAppVersions == nil {
		return "unknown"
	}
	return fmt.Sprint(*e.Result.NumAppVersions)
}

// print a summary of the experiment
func (e *experiment) printSummary(w *tabwriter.Writer) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "Experiment summary\t")
	fmt.Fprintln(w, "--------------------------\t-----")
	fmt.Fprintln(w, "Num app versions \t"+e.numAppVersions())
	fmt.Fprintln(w, "Winner \t"+e.winner())
	fmt.Fprintln(w, "Testing pattern \t"+e.testingPatternString())
	fmt.Fprintln(w, "Experiment completed \t"+strconv.FormatBool(e.completed()))
	fmt.Fprintln(w, "Experiment failed \t"+strconv.FormatBool(!e.noFailure()))
	fmt.Fprintln(w, "Number of completed tasks \t"+e.numCompletedTasksString())
	fmt.Fprintln(w, "")
	w.Flush()
}

// // printWinnerAssessment prints the winning version in the experiment into d's description buffer.
// // If winner assessment is unavailable for the underlying experiment, this method will indicate likewise.
// func (d *Result) printWinnerAssessment() *Result {
// 	if d.err != nil {
// 		return d
// 	}
// 	if a := d.experiment.Status.Analysis; a != nil {
// 		if w := a.WinnerAssessment; w != nil {
// 			d.description.WriteString("\n****** Winner Assessment ******\n")
// 			var explanation string = ""
// 			switch d.experiment.Spec.Strategy.TestingPattern {
// 			case v2alpha2.TestingPatternCanary:
// 				explanation = "> If the candidate version satisfies the experiment objectives, then it is the winner.\n> Otherwise, if the baseline version satisfies the experiment objectives, it is the winner.\n> Otherwise, there is no winner.\n"
// 			case v2alpha2.TestingPatternConformance:
// 				explanation = "> If the version being validated; i.e., the baseline version, satisfies the experiment objectives, it is the winner.\n> Otherwise, there is no winner.\n"
// 			default:
// 				explanation = ""
// 			}
// 			d.description.WriteString(explanation)
// 			if d.experiment.Spec.Strategy.TestingPattern != v2alpha2.TestingPatternConformance && d.experiment.Spec.VersionInfo != nil {
// 				versions := []string{d.experiment.Spec.VersionInfo.Baseline.Name}
// 				for i := 0; i < len(d.experiment.Spec.VersionInfo.Candidates); i++ {
// 					versions = append(versions, d.experiment.Spec.VersionInfo.Candidates[i].Name)
// 				}
// 				d.description.WriteString(fmt.Sprintf("App versions in this experiment: %s\n", versions))
// 			}
// 			if w.Data.WinnerFound {
// 				d.description.WriteString(fmt.Sprintf("Winning version: %s\n", *w.Data.Winner))
// 			} else {
// 				d.description.WriteString("Winning version: not found\n")
// 			}

// 			if d.experiment.Spec.Strategy.TestingPattern != v2alpha2.TestingPatternConformance &&
// 				d.experiment.Status.VersionRecommendedForPromotion != nil {
// 				d.description.WriteString(fmt.Sprintf("Version recommended for promotion: %s\n", *d.experiment.Status.VersionRecommendedForPromotion))
// 			}
// 		}
// 	}
// 	return d
// }

// // printRewardAssessment prints a matrix of values for each reward-version pair.
// // Rows correspond to experiment rewards. Columns correspond to versions.
// // The current "best" version for each reward is denoted with a "*".
// func (d *Result) printRewardAssessment() *Result {
// 	if d.err != nil ||
// 		d.experiment.Status.Analysis == nil ||
// 		d.experiment.Status.Analysis.VersionAssessments == nil ||
// 		d.experiment.Spec.Criteria == nil ||
// 		len(d.experiment.Spec.Criteria.Rewards) == 0 {
// 		return d
// 	}

// 	d.description.WriteString("\n****** Reward Assessment ******\n")
// 	d.description.WriteString("> Identifies values of reward metrics for each version. The best version is marked with a '*'.\n")
// 	table := tablewriter.NewWriter(&d.description)
// 	table.SetRowLine(true)
// 	versions := d.experiment.GetVersions()
// 	table.SetHeader(append([]string{"Reward"}, versions...))
// 	for _, reward := range d.experiment.Spec.Criteria.Rewards {
// 		row := []string{expr.StringifyReward(reward)}
// 		table.Append(append(row, d.experiment.GetAnnotatedMetricStrs(reward)...))
// 	}
// 	table.Render()

// 	return d
// }

// // printObjectiveAssessment prints a matrix of boolean values into d's description buffer.
// // Rows correspond to experiment objectives, columns correspond to versions, and entry [i, j] indicates if objective i is satisfied by version j.
// // Objective assessments are printed in the same sequence as in the experiment's spec.criteria.objectives section.
// // If objective assessments are unavailable for the underlying experiment, this method will indicate likewise.
// func (d *Result) printObjectiveAssessment() *Result {
// 	if d.err != nil {
// 		return d
// 	}
// 	if a := d.experiment.Status.Analysis; a != nil {
// 		if v := a.VersionAssessments; v != nil {
// 			d.description.WriteString("\n****** Objective Assessment ******\n")
// 			d.description.WriteString("> Whether objectives specified in the experiment are satisfied by versions.\n")
// 			d.description.WriteString("> This assessment is based on last known metric values for each version.\n")
// 			table := tablewriter.NewWriter(&d.description)
// 			table.SetRowLine(true)
// 			versions := d.experiment.GetVersions()
// 			table.SetHeader(append([]string{"Metric", "Condition"}, versions...))
// 			for i, objective := range d.experiment.Spec.Criteria.Objectives {
// 				row := []string{objective.Metric, expr.ConditionFromObjective(objective)}
// 				table.Append(append(row, d.experiment.GetSatisfyStrs(i)...))
// 			}
// 			table.Render()
// 		}
// 	}
// 	return d
// }

// // printVersionAssessment prints how each version is performing with respect to experiment criteria into d's description buffer. This method invokes printObjectiveAssessment under the covers.
// func (d *Result) printVersionAssessment() *Result {
// 	if d.err != nil {
// 		return d
// 	}
// 	if c := d.experiment.Spec.Criteria; c != nil && len(c.Objectives) > 0 {
// 		d.printObjectiveAssessment()
// 	}
// 	return d
// }

// // printMetrics prints a matrix of (decimal) metric values into d's description buffer.
// // Rows correspond to experiment metrics, columns correspond to versions, and entry [i, j] indicates the value of metric i for version j.
// // Metrics are printed in the same sequence as in the experiment's status.metrics section.
// // If metrics are unavailable for the underlying experiment, this method will indicate likewise.
// func (d *Result) printMetrics() *Result {
// 	if d.err != nil {
// 		return d
// 	}
// 	if a := d.experiment.Status.Analysis; a != nil {
// 		if v := a.AggregatedMetrics; v != nil {
// 			d.description.WriteString("\n****** Metrics Assessment ******\n")
// 			d.description.WriteString("> Last known metric values for each version.\n")
// 			table := tablewriter.NewWriter(&d.description)
// 			table.SetRowLine(true)
// 			versions := d.experiment.GetVersions()
// 			table.SetHeader(append([]string{"Metric"}, versions...))
// 			for _, metricInfo := range d.experiment.Status.Metrics {
// 				row := []string{expr.GetMetricNameAndUnits(metricInfo)}
// 				table.Append(append(row, d.experiment.GetMetricStrs(metricInfo.Name)...))
// 			}
// 			table.Render()
// 		}
// 	}
// 	return d
// }

// // PrintAnalysis prints the progress of the iter8 experiment, winner assessment, version assessment, and metrics.
// func (d *Result) PrintAnalysis() *Result {
// 	if d.err != nil {
// 		return d
// 	}
// 	d.printProgress()
// 	if d.experiment.Started() {
// 		d.printWinnerAssessment().
// 			printRewardAssessment().
// 			printVersionAssessment().
// 			printMetrics()
// 	}
// 	if d.err == nil {
// 		fmt.Fprintln(os.Stdout, d.description.String())
// 	}
// 	return d
// }
