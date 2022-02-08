package basecli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/base"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli/values"
)

func TestMockQuickStartWithoutSLOs(t *testing.T) {
	// mock
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))
	defer httpmock.Deactivate()

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// without SLOs first

	// gen and run exp
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=https://example.com", "duration=2s"},
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
	// mock
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))
	defer httpmock.Deactivate()

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// with SLOs next
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=https://example.com", "duration=2s"},
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
	// mock
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))
	defer httpmock.Deactivate()

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// with bad SLOs
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=https://example.com", "duration=2s"},
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
	// mock
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))
	defer httpmock.Deactivate()

	// get into the experiment chart folder
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))

	// with SLOs and percentiles also
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=https://example.com", "duration=2s"},
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
	// mock
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))
	defer httpmock.Deactivate()

	// dry run
	os.Chdir(base.CompletePath("../", "hub/load-test-http"))
	Dry = true
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=https://example.com", "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	err := runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")
}
func TestDryRun(t *testing.T) {
	// mock
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))
	defer httpmock.Deactivate()

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
		StringValues: []string{"url=https://example.com", "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	err = runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")
}
