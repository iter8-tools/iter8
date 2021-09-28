package describe

import (
	"fmt"
	"testing"

	"github.com/iter8-tools/etc3/iter8ctl/utils"
	"github.com/stretchr/testify/assert"
)

/* Tests */

func TestPrintProgress(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.printProgress()
		assert.NoError(t, d.Error())
	}
}

func TestPrintWinnerAssessment(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.printWinnerAssessment()
		assert.NoError(t, d.Error())
	}
}

func TestPrintObjectiveAssessment(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.printObjectiveAssessment()
		assert.NoError(t, d.Error())
	}
}

func TestPrintVersionAssessment(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.printVersionAssessment()
		assert.NoError(t, d.Error())
	}
}

func TestPrintMetrics(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.printMetrics()
		assert.NoError(t, d.Error())
	}
}

func TestPrintRewardAssessments(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.printRewardAssessment()
		assert.NoError(t, d.Error())
	}
}

func TestPrintAnalysis(t *testing.T) {
	for i := 1; i <= 12; i++ {
		d := Builder().FromFile(utils.CompletePath("../", fmt.Sprintf("testdata/experiment%v.yaml", i)))
		d.PrintAnalysis()
		assert.NoError(t, d.Error())
	}
}
