package experiment

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/iter8ctl/utils"
	tasks "github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// getExp is a helper function for extracting an experiment object from experiment filenamePrefix
// filePath is relative to testdata folder
func getExp(filenamePrefix string) (*Experiment, error) {
	experimentFilepath := utils.CompletePath("../", fmt.Sprintf("testdata/%s.yaml", filenamePrefix))
	expBytes, err := ioutil.ReadFile(experimentFilepath)
	if err != nil {
		return nil, err
	}

	exp := &Experiment{}
	err = yaml.Unmarshal(expBytes, exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

type test struct {
	name                   string // name of this test
	started                bool
	exp                    *Experiment
	errorRates, fakeMetric []string
	satisfyStrs, fakeObj   []string
}

var fakeValStrs = []string{"unavailable", "unavailable"}

var satisfyStrs = []string{"true", "true"}

var errorRateStrs = []string{"0.000", "0.000"}

// table driven tests
var tests = []test{
	{name: "experiment1", started: false, errorRates: []string{}, fakeMetric: []string{}, satisfyStrs: []string{}, fakeObj: []string{}},
	{name: "experiment2", started: false, errorRates: []string{}, fakeMetric: []string{}, satisfyStrs: []string{}, fakeObj: []string{}},
	{name: "experiment3", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment4", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment5", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment6", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment7", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment8", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
	{name: "experiment9", started: true, errorRates: errorRateStrs, fakeMetric: fakeValStrs, satisfyStrs: satisfyStrs, fakeObj: fakeValStrs},
}

func init() {
	for i := 0; i < len(tests); i++ {
		e, err := getExp(tests[i].name)
		if err == nil {
			tests[i].exp = e
		} else {
			fmt.Println("Unable to extract experiment objects from files")
			os.Exit(1)
		}
	}
}

/* Tests */

func TestExperiment(t *testing.T) {
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// test Started()
			assert.Equal(t, tc.started, tc.exp.Started())
			// test GetVersions()
			if tc.exp.Started() {
				assert.Equal(t, []string{"default", "canary"}, tc.exp.GetVersions())
			} else {
				assert.Equal(t, []string([]string(nil)), tc.exp.GetVersions())
			}
			// test GetMetricStrs(...)
			assert.Equal(t, tc.errorRates, tc.exp.GetMetricStrs("error-rate"))
			assert.Equal(t, tc.fakeMetric, tc.exp.GetMetricStrs("fake-metric"))
			// test GetSatisfyStrs()
			assert.Equal(t, tc.satisfyStrs, tc.exp.GetSatisfyStrs(0))
			assert.Equal(t, tc.fakeObj, tc.exp.GetSatisfyStrs(10))
		})
	}
}

func TestGetMetricNameAndUnits(t *testing.T) {
	metricNameAndUnits := [4]string{"95th-percentile-tail-latency (milliseconds)", "mean-latency (milliseconds)", "error-rate", "request-count"}
	mnu := [4]string{}
	for i := 0; i < 4; i++ {
		mnu[i] = GetMetricNameAndUnits(tests[2].exp.Status.Metrics[i])
	}
	assert.Equal(t, metricNameAndUnits, mnu)
}

func TestStringifyObjective(t *testing.T) {
	objectives := [2]string{"mean-latency <= 1000.000", "error-rate <= 0.010"}
	objs := [2]string{}
	for i := 0; i < 2; i++ {
		objs[i] = StringifyObjective(tests[2].exp.Spec.Criteria.Objectives[i])
	}
	assert.Equal(t, objectives, objs)
}

func TestStringifyReward(t *testing.T) {
	assert.Equal(t,
		"reward (lower better)",
		StringifyReward(v2alpha2.Reward{Metric: "reward", PreferredDirection: "Lower"}))
	assert.Equal(t,
		"reward (higher better)",
		StringifyReward(v2alpha2.Reward{Metric: "reward", PreferredDirection: "High"}))
}

func TestGetAnnotatedMetricStrs(t *testing.T) {
	e, err := getExp("experiment12")
	assert.NoError(t, err)
	assert.Equal(t,
		[]string{"5.030", "24.454 *"},
		e.GetAnnotatedMetricStrs(v2alpha2.Reward{Metric: "books-purchased", PreferredDirection: "High"}))
}

func TestAssertComplete(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").WithCondition(
		v2alpha2.ExperimentConditionExperimentCompleted,
		corev1.ConditionTrue,
		"experiment is over",
		"",
	).Build()

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{Completed})

	assert.NoError(t, err)
}

func TestAssertInComplete(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").WithCondition(
		v2alpha2.ExperimentConditionExperimentCompleted,
		corev1.ConditionFalse,
		"experiment is not over",
		"",
	).Build()

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{Completed})

	assert.Error(t, err)
}

func TestAssertWinnerFound(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = &v2alpha2.WinnerAssessmentAnalysis{
		Data: v2alpha2.WinnerAssessmentData{
			WinnerFound: true,
			Winner:      tasks.StringPointer("the best"),
		},
	}

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.NoError(t, err)
}

func TestAssertNoWinnerFound(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = &v2alpha2.WinnerAssessmentAnalysis{
		AnalysisMetaData: v2alpha2.AnalysisMetaData{},
		Data: v2alpha2.WinnerAssessmentData{
			WinnerFound: false,
		},
	}

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.Error(t, err)
}

func TestAssertNoWinnerFound2(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = nil

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.Error(t, err)
}

func TestAssertNoWinnerFound3(t *testing.T) {
	exp := v2alpha2.NewExperiment("test", "test").Build()
	exp.Status.Analysis = nil

	err := (&Experiment{
		*exp,
	}).Assert([]ConditionType{WinnerFound})

	assert.Error(t, err)
}

/* Examples */

func ExampleGetMetricNameAndUnits() {
	u := "inches"
	mi := v2alpha2.MetricInfo{
		Name: "height",
		MetricObj: v2alpha2.Metric{
			Spec: v2alpha2.MetricSpec{
				Units: &u,
			},
		},
	}
	met := GetMetricNameAndUnits(mi)
	fmt.Println(met)
	// Output: height (inches)
}

func ExampleGetMetricNameAndUnits_unitless() {
	mi := v2alpha2.MetricInfo{
		Name:      "weight",
		MetricObj: v2alpha2.Metric{},
	}
	met := GetMetricNameAndUnits(mi)
	fmt.Println(met)
	// Output: weight
}

func ExampleStringifyObjective_upperlimit() {
	q := resource.MustParse("0.01")
	obj := v2alpha2.Objective{
		Metric:     "error-rate",
		UpperLimit: &q,
		LowerLimit: nil,
	}
	str := StringifyObjective(obj)
	fmt.Println(str)
	// Output: error-rate <= 0.010
}

func ExampleStringifyObjective_lowerlimit() {
	q := resource.MustParse("0.998")
	obj := v2alpha2.Objective{
		Metric:     "accuracy",
		UpperLimit: nil,
		LowerLimit: &q,
	}
	str := StringifyObjective(obj)
	fmt.Println(str)
	// Output: 0.998 <= accuracy
}

func ExampleStringifyObjective_upperandlower() {
	q1 := resource.MustParse("6.998")
	q2 := resource.MustParse("7.012")
	obj := v2alpha2.Objective{
		Metric:     "pH level",
		UpperLimit: &q2,
		LowerLimit: &q1,
	}
	str := StringifyObjective(obj)
	fmt.Println(str)
	// Output: 6.998 <= pH level <= 7.012
}

func ExampleExperiment_GetMetricStr() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 'mean-latency' metric for 'canary' version
	met := exp.GetMetricStr("mean-latency", "canary")
	fmt.Println(met)
	// output: 197.500
}

func ExampleExperiment_GetMetricStr_unavailable1() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 'fake' metric for 'default' version
	met := exp.GetMetricStr("fake", "default")
	fmt.Println(met)
	// output: unavailable
}

func ExampleExperiment_GetMetricStr_unavailable2() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 'mean-latency' metric for 'perfect' version
	met := exp.GetMetricStr("mean-latency", "perfect")
	fmt.Println(met)
	// output: unavailable
}

func ExampleExperiment_GetMetricStrs() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 'mean-latency' metric for all versions ('default' and 'canary')
	mets := exp.GetMetricStrs("mean-latency")
	fmt.Println(mets)
	// output: [228.788 197.500]
}

func ExampleExperiment_GetMetricStrs_unavailable() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 'fake' metric for all versions ('default' and 'canary')
	mets := exp.GetMetricStrs("fake")
	fmt.Println(mets)
	// output: [unavailable unavailable]
}

func ExampleExperiment_GetSatisfyStr() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 2nd objective for 'canary' version
	obj := exp.GetSatisfyStr(1, "canary")
	fmt.Println(obj)
	// output: true
}

func ExampleExperiment_GetSatisfyStr_unavailable1() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 3rd objective for 'default' version
	// This experiment has only two objectives, so this value is unavailable
	obj := exp.GetSatisfyStr(2, "default")
	fmt.Println(obj)
	// output: unavailable
}

func ExampleExperiment_GetSatisfyStr_unavailable2() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment2.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of the 2nd objective for 'perfect' version
	obj := exp.GetSatisfyStr(1, "perfect")
	fmt.Println(obj)
	// output: unavailable
}

func ExampleExperiment_GetSatisfyStrs() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of objective indicators for the 2nd objective for all versions ('default' and 'canary')
	objs := exp.GetSatisfyStrs(1)
	fmt.Println(objs)
	// output: [true true]
}

func ExampleExperiment_GetSatisfyStrs_unavailable1() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment2.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of objective indicators for the 2nd objective for all versions ('default' and 'canary')
	// This experiment does not have versionInfo as part of its spec section, so there are no versions
	objs := exp.GetSatisfyStrs(1)
	fmt.Println(objs)
	// output: []
}

func ExampleExperiment_GetSatisfyStrs_unavailable2() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of objective indicators for the 3nd objective for all versions ('default' and 'canary')
	// The experiment has only two objectives, so these values are unavailable.
	objs := exp.GetSatisfyStrs(2)
	fmt.Println(objs)
	// output: [unavailable unavailable]
}

func ExampleExperiment_GetVersions() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of objective indicators for the 2nd objective for all versions ('default' and 'canary')
	versions := exp.GetVersions()
	fmt.Println(versions)
	// output: [default canary]
}

func ExampleExperiment_GetVersions_empty() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment2.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	// Get value of objective indicators for the 2nd objective for all versions ('default' and 'canary')
	// This experiment does not have versionInfo as part of its spec section, so there are no versions
	versions := exp.GetVersions()
	fmt.Println(versions)
	// output: []
}

func ExampleExperiment_Started_true() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment3.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	started := exp.Started()
	fmt.Println(started)
	// output: true
}

func ExampleExperiment_Started_false() {
	// Read in an experiment object from the testdata folder
	filePath := utils.CompletePath("../testdata", "experiment2.yaml")
	buf, _ := ioutil.ReadFile(filePath)
	exp := &Experiment{}
	yaml.Unmarshal(buf, exp)
	started := exp.Started()
	fmt.Println(started)
	// output: false
}
