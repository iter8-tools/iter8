package action

import (
	"fmt"
	"strings"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/iter8-tools/iter8/report"
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
	driver.KubeDriver
}

func NewReportOpts() *ReportOpts {
	return &ReportOpts{
		RunOpts: RunOpts{
			RunDir: ".",
		},
	}
}

func (rOpts *ReportOpts) LocalRun() error {
	return rOpts.Run(&driver.FileDriver{
		RunDir: rOpts.RunDir,
	})
}

func (rOpts *ReportOpts) KubeRun() error {
	if err := rOpts.KubeDriver.Init(); err != nil {
		return err
	}
	return rOpts.Run(rOpts)
}

func (rOpts *ReportOpts) Run(eio base.Driver) error {
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
			return reporter.Gen()
		case HTMLOutputFormatKey:
			reporter := report.HTMLReporter{
				Reporter: &report.Reporter{
					Experiment: e,
				},
			}
			return reporter.Gen()
		default:
			e := fmt.Errorf("unsupported report format %v", rOpts.OutputFormat)
			log.Logger.Error(e)
			return e
		}
	}
}
