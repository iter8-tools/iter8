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

func TestMockQuickStartWithoutSLOs(t *testing.T) {
	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// without SLOs first

	// gen and run exp
	GenOptions.Values = append(GenOptions.Values, "url=https://example.com")
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
	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// with SLOs next
	GenOptions.Values = append(GenOptions.Values, "url=https://example.com", "SLOs.error-rate=0", "SLOs.latency-mean=100", "duration=2s")
	GenOptions.ValueFiles = append(GenOptions.ValueFiles, base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml"))
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
	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// with bad SLOs
	GenOptions = values.Options{
		Values:     []string{"url=https://example.com", "SLOs.error-rate=0", "SLOs.latency-mean=100", "SLOs.latency-p95=0.00001", "duration=2s"},
		ValueFiles: []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
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
	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// with SLOs and percentiles also
	GenOptions = values.Options{
		Values: []string{"url=https://example.com", "SLOs.error-rate=0", "SLOs.latency-mean=100", "SLOs.latency-p50=100"},
	}

	GenOptions.ValueFiles = append(GenOptions.ValueFiles, base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml"))
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// assert
	AssertOptions = AssertOptionsType{
		Conds:   []string{Completed, NoFailure, SLOs, "slosby=0"},
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
	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// dry run
	os.Chdir(base.CompletePath("../", "hub/load-test"))
	Dry = true
	GenOptions = values.Options{
		Values: []string{"url=https://example.com"},
	}
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")

}
func TestDryRun(t *testing.T) {
	dir, _ := ioutil.TempDir("", "iter8-test")
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	hubFolder = "load-test"
	// hub
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// dry run
	os.Chdir(path.Join(dir, hubFolder))
	Dry = true
	GenOptions = values.Options{
		Values: []string{"url=https://example.com"},
	}
	err = runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")

}
