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
)

func TestMockQuickStartWithoutSLOs(t *testing.T) {
	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test"))

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
	os.Chdir(base.CompletePath("../", "hub/load-test"))

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// with SLOs next
	GenOptions.Values = append(GenOptions.Values, "url=https://example.com", "SLOs.error-rate=0", "SLOs.mean-latency=100")
	GenOptions.ValueFiles = append(GenOptions.ValueFiles, base.CompletePath("../", "testdata/percentileandslos/values.yaml"))
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

func TestMockQuickStartWithSLOsAndPercentiles(t *testing.T) {
	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test"))

	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// with SLOs and percentiles also
	GenOptions.Values = append(GenOptions.Values, "url=https://example.com", "SLOs.error-rate=0", "SLOs.mean-latency=100", "SLOs.p50=100")
	GenOptions.ValueFiles = append(GenOptions.ValueFiles, base.CompletePath("../", "testdata/percentileandslos/values.yaml"))
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// assert
	AssertOptions = AssertOptionsType{
		Conds:   []string{Completed, NoFailure, SLOs},
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

func TestDryRun(t *testing.T) {
	dir, _ := ioutil.TempDir("", "iter8-test")
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	// ToDo: Fix this location
	os.Setenv("ITER8HUB", "github.com/sriumcp/iter8.git?ref=grpc//hub/")
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
	GenOptions.Values = append(GenOptions.Values, "url=https://example.com")
	err = runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")
}
