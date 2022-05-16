package base

import (
	"errors"
	"io/ioutil"
	"path"

	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

//FileDriver enables reading and writing experiment spec and result files
type FileDriver struct {
	// RunDir is the directory where the experiment.yaml file is to be found
	RunDir string
}

// ReadSpec reads experiment spec from file
func (f *FileDriver) ReadSpec() (ExperimentSpec, error) {
	b, err := ioutil.ReadFile(path.Join(f.RunDir, ExperimentSpecPath))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, errors.New("unable to read experiment spec")
	}
	return SpecFromBytes(b)
}

// // ReadMetricsSpec reads metrics spec from file
// func (f *FileDriver) ReadMetricsSpec(provider string) (*template.Template, error) {
// 	// b, err := ioutil.ReadFile(path.Join(f.RunDir, provider, ExperimentMetricsPathSuffix))
// 	// if err != nil {
// 	// 	log.Logger.WithStackTrace(err.Error()).Error("unable to read metrics spec")
// 	// 	return nil, errors.New("unable to read metrics spec")
// 	// }
// 	// return MetricsSpecFromBytes(b)

// 	return nil, nil
// }

// ReadResult reads experiment result from file
func (f *FileDriver) ReadResult() (*ExperimentResult, error) {
	b, err := ioutil.ReadFile(path.Join(f.RunDir, ExperimentResultPath))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, errors.New("unable to read experiment result")
	}
	return ResultFromBytes(b)
}

// WriteResult writes experiment result to file
func (f *FileDriver) WriteResult(res *ExperimentResult) error {
	rBytes, _ := yaml.Marshal(res)
	err := ioutil.WriteFile(path.Join(f.RunDir, ExperimentResultPath), rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}
