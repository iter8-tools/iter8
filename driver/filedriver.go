package driver

import (
	"errors"
	"os"
	"path"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

// FileDriver enables reading and writing experiment spec and result files
type FileDriver struct {
	// RunDir is the directory where the experiment.yaml file is to be found
	RunDir string
}

// Read the experiment
func (f *FileDriver) Read() (*base.Experiment, error) {
	b, err := os.ReadFile(path.Join(f.RunDir, ExperimentPath))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment")
		return nil, errors.New("unable to read experiment")
	}
	return ExperimentFromBytes(b)
}

// Write the experiment
func (f *FileDriver) Write(exp *base.Experiment) error {
	b, _ := yaml.Marshal(exp)
	err := os.WriteFile(path.Join(f.RunDir, ExperimentPath), b, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment")
		return errors.New("unable to write experiment")
	}
	return nil
}

// GetRevision is undefined for file drivers
func (f *FileDriver) GetRevision() int {
	return 0
}
