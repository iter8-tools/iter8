package basecli

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestMockQuickStart(t *testing.T) {
	dir, _ := ioutil.TempDir("", "iter8-test")
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	// Todo: fix location below
	os.Setenv("ITER8HUB", "github.com/sriumcp/iter8.git?ref=chart//hub/")
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

	// gen exp
	os.Chdir(path.Join(dir, hubFolder))
	GenOptions.Values = append(GenOptions.Values, "url=https://example.com")
	err = expCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// run experiment
	err = runCmd.RunE(nil, nil)
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
	err = reportCmd.RunE(nil, nil)
	assert.NoError(t, err)

	// report HTML
	ReportOptions = ReportOptionsType{
		OutputFormat: HTMLOutputFormatKey,
	}
	err = reportCmd.RunE(nil, nil)
	assert.NoError(t, err)
}
