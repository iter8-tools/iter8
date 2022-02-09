package basecli

import (
	"fmt"
	"io"
	"path"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli/values"
)

func TestMockQuickStartWithoutSLOs(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// gen and run exp
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	chartPath = base.CompletePath("../", "hub/load-test-http")
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
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// with SLOs next
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{"SLOs.error-rate=0", "SLOs.latency-mean=100"},
		FileValues:   []string{},
	}
	chartPath = base.CompletePath("../", "hub/load-test-http")
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
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// with bad SLOs
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{"SLOs.error-rate=0", "SLOs.latency-mean=100", "SLOs.latency-p95=0.00001"},
	}
	chartPath = base.CompletePath("../", "hub/load-test-http")
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
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// with SLOs and percentiles also
	GenOptions = values.Options{
		ValueFiles:   []string{base.CompletePath("../", "testdata/percentileandslos/load-test-http-values.yaml")},
		StringValues: []string{"url=" + testURL, "duration=2s"},
		Values:       []string{"SLOs.error-count=0", "SLOs.latency-mean=100", "SLOs.latency-p50=100"},
		FileValues:   []string{},
	}
	chartPath = base.CompletePath("../", "hub/load-test-http")
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

func TestDryRun(t *testing.T) {
	chartName = "load-test-http"
	// hub
	err := hubCmd.RunE(nil, nil)
	assert.NoError(t, err)

	dry = true
	GenOptions = values.Options{
		ValueFiles:   []string{},
		StringValues: []string{"url=https://example.com", "duration=2s"},
		Values:       []string{},
		FileValues:   []string{},
	}
	chartPath = path.Join(destDir, chartName)
	err = runCmd.RunE(nil, nil)
	assert.NoError(t, err)
	assert.FileExists(t, "experiment.yaml")
}
