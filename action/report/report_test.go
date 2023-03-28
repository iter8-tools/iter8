package report

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
)

func TestReportText(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	_ = copyFileToPwd(t, base.CompletePath("../../", "testdata/assertinputs/experiment.yaml"))

	fd := driver.FileDriver{
		RunDir: ".",
	}
	exp, err := base.BuildExperiment(&fd)
	assert.NoError(t, err)
	reporter := TextReporter{
		Reporter: &Reporter{
			Experiment: exp,
		},
	}
	err = reporter.Gen(os.Stdout)
	assert.NoError(t, err)
}

func TestReportTextWithLowerSLOs(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	_ = copyFileToPwd(t, base.CompletePath("../../", "testdata/assertinputs/experimentWithLowerSLOs.yaml"))
	_ = os.Rename("experimentWithLowerSLOs.yaml", "experiment.yaml")

	fd := driver.FileDriver{
		RunDir: ".",
	}
	exp, err := base.BuildExperiment(&fd)
	assert.NoError(t, err)
	reporter := TextReporter{
		Reporter: &Reporter{
			Experiment: exp,
		},
	}
	err = reporter.Gen(os.Stdout)
	assert.NoError(t, err)
}

func TestReportHTMLWithLowerSLOs(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	_ = copyFileToPwd(t, base.CompletePath("../../", "testdata/assertinputs/experimentWithLowerSLOs.yaml"))
	_ = os.Rename("experimentWithLowerSLOs.yaml", "experiment.yaml")
	fd := driver.FileDriver{
		RunDir: ".",
	}
	exp, err := base.BuildExperiment(&fd)
	assert.NoError(t, err)
	reporter := HTMLReporter{
		Reporter: &Reporter{
			Experiment: exp,
		},
	}
	err = reporter.Gen(os.Stdout)
	assert.NoError(t, err)
}
