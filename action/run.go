package action

type Run struct {
	RunDir string
}

func NewRun() *Run {
	return &Run{
		RunDir: ".",
	}
}

func (runner *Run) getFileOps() *fileOps {
	return &fileOps{
		RunDir: runner.RunDir,
	}
}

func (runner *Run) RunLocal() error {
	return runExperiment(runner.getFileOps())
}
