package basecli

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	testName    = "example"
	testPort    = "8000"
	testHost    = "localhost"
	testAddress = testHost + ":" + testPort
	testPath    = "/"
	testURL     = "http://" + testHost + testPath
)

func TestMockQuickStartWithoutSLOs(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	// Exact URL match
	httpmock.RegisterResponder("GET", testURL,
		httpmock.NewStringResponder(200, `all good`))

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// gen and run exp
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// assert
	AssertOptions = AssertOptionsType{
		Conds:   []string{Completed, NoFailure, SLOs},
		Timeout: 0,
	}
	err = assertCmd.RunE(nil, nil)
	assert.NoError(t, err)
}

func TestMockQuickStartWithSLOs(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	// Exact URL match
	httpmock.RegisterResponder("GET", testURL,
		httpmock.NewStringResponder(200, `all good`))

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// with SLOs next
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{"SLOs.error-rate=0", "SLOs.latency-mean=100"},
		FileValues:   []string{},
	}
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// assert
	AssertOptions = AssertOptionsType{
		Conds:   []string{Completed, NoFailure, SLOs},
		Timeout: 0,
	}
	err = assertCmd.RunE(nil, nil)
	assert.NoError(t, err)
}

func TestMockQuickStartWithBadSLOs(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	// Exact URL match
	httpmock.RegisterResponder("GET", testURL,
		httpmock.NewStringResponder(200, `all good`))

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// with bad SLOs
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{"SLOs.error-rate=0", "SLOs.latency-mean=100", "SLOs.latency-p95=0.00001"},
	}
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// assert
	AssertOptions = AssertOptionsType{
		Conds:   []string{Completed, NoFailure, SLOs},
		Timeout: 5,
	}

	exp, _ := Build(true, &FileExpIO{})
	allGood, err := exp.Assert(AssertOptions.Conds, AssertOptions.Timeout)
	assert.NoError(t, err)
	assert.False(t, allGood)
}

func TestMockQuickStartWithSLOsAndPercentiles(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	// Exact URL match
	httpmock.RegisterResponder("GET", testURL,
		httpmock.NewStringResponder(200, `all good`))

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// with SLOs and percentiles also
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{"SLOs.error-count=0", "SLOs.latency-mean=100", "SLOs.latency-p50=100"},
		FileValues:   []string{},
	}

	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// assert
	AssertOptions = AssertOptionsType{
		Conds:   []string{Completed, NoFailure, SLOsByPrefix + "=0"},
		Timeout: 0,
	}
	err = assertCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// report text
	ReportOptions = ReportOptionsType{
		OutputFormat: TextOutputFormatKey,
	}
	reportCmd.SetOut(io.Discard)
	err = reportCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// report HTML
	ReportOptions = ReportOptionsType{
		OutputFormat: HTMLOutputFormatKey,
	}
	err = reportCmd.RunE(nil, nil)
	assert.NoError(t, err)
}

func TestDryRunLocal(t *testing.T) {
	// dry run
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))
	Dry = true
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")
}
func TestDryRun(t *testing.T) {
	dir, _ := ioutil.TempDir("", "iter8-test")
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	hubFolder = "load-test-http"
	// hub
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// dry run
	os.Chdir(path.Join(dir, hubFolder))
	Dry = true
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	err = runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")
}
