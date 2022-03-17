package action

import (
	"fmt"
	"io"
	"strings"

	"github.com/iter8-tools/iter8/action/report"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

const (
	// TextOutputFormat is the output format used to create text output
	TextOutputFormatKey = "text"

	// HTMLOutputFormat is the output format used to create html output
	HTMLOutputFormatKey = "html"
)

type ReportOpts struct {
	// OutputFormat holds the output format to be used by report
	OutputFormat string
	// applicable only for local experiments
	RunOpts
	// applicable only for Kubernetes experiments
	*driver.KubeDriver
}

func NewReportOpts(kd *driver.KubeDriver) *ReportOpts {
	return &ReportOpts{
		RunOpts: RunOpts{
			RunDir: ".",
		},
		OutputFormat: TextOutputFormatKey,
		KubeDriver:   kd,
	}
}

func (rOpts *ReportOpts) LocalRun(out io.Writer) error {
	return rOpts.Run(&driver.FileDriver{
		RunDir: rOpts.RunDir,
	}, out)
}

func (rOpts *ReportOpts) KubeRun(out io.Writer) error {
	if err := rOpts.KubeDriver.Init(); err != nil {
		return err
	}
	return rOpts.Run(rOpts, out)
}

func (rOpts *ReportOpts) Run(eio base.Driver, out io.Writer) error {
	if e, err := base.BuildExperiment(true, eio); err != nil {
		return err
	} else {
		switch strings.ToLower(rOpts.OutputFormat) {
		case TextOutputFormatKey:
			reporter := report.TextReporter{
				Reporter: &report.Reporter{
					Experiment: e,
				},
			}
			return reporter.Gen(out)
		case HTMLOutputFormatKey:
			reporter := report.HTMLReporter{
				Reporter: &report.Reporter{
					Experiment: e,
				},
			}
			return reporter.Gen(out)
		default:
			e := fmt.Errorf("unsupported report format %v", rOpts.OutputFormat)
			log.Logger.Error(e)
			return e
		}
	}
}
