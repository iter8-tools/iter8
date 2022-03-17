package report

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
)

func TestReportText(t *testing.T) {

	fd := driver.FileDriver{
		RunDir: base.CompletePath("../../", "testdata/assertinputs"),
	}
	exp, err := base.BuildExperiment(true, &fd)
	assert.NoError(t, err)
	reporter := TextReporter{
		Reporter: &Reporter{
			Experiment: exp,
		},
	}
	err = reporter.Gen(os.Stdout)
	assert.NoError(t, err)
}

func TestReportHTML(t *testing.T) {

	fd := driver.FileDriver{
		RunDir: base.CompletePath("../../", "testdata/assertinputs"),
	}
	exp, err := base.BuildExperiment(true, &fd)
	assert.NoError(t, err)
	reporter := HTMLReporter{
		Reporter: &Reporter{
			Experiment: exp,
		},
	}
	err = reporter.Gen(os.Stdout)
	assert.NoError(t, err)
}
