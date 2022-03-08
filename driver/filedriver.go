package driver

import (
	"errors"
	"io/ioutil"
	"path"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

//FileDriver enables reading and writing experiment spec and result files
type FileDriver struct {
	RunDir string
}

// SpecFromBytes reads experiment spec from bytes
func SpecFromBytes(b []byte) (base.ExperimentSpec, error) {
	e := base.ExperimentSpec{}
	err := yaml.Unmarshal(b, &e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment spec")
		return nil, err
	}
	return e, err
}

// ResultFromBytes reads experiment result from bytes
func ResultFromBytes(b []byte) (*base.ExperimentResult, error) {
	r := &base.ExperimentResult{}
	err := yaml.Unmarshal(b, r)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment result")
		return nil, err
	}
	return r, err
}

// ReadSpec reads experiment spec from file
func (f *FileDriver) ReadSpec() (base.ExperimentSpec, error) {
	b, err := ioutil.ReadFile(path.Join(f.RunDir, ExperimentSpecPath))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, errors.New("unable to read experiment spec")
	}
	return SpecFromBytes(b)
}

// ReadResult reads experiment result from file
func (f *FileDriver) ReadResult() (*base.ExperimentResult, error) {
	b, err := ioutil.ReadFile(path.Join(f.RunDir, ExperimentResultPath))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, errors.New("unable to read experiment result")
	}
	return ResultFromBytes(b)
}

// WriteResult writes experiment result to file
func (f *FileDriver) WriteResult(res *base.ExperimentResult) error {
	rBytes, _ := yaml.Marshal(res)
	err := ioutil.WriteFile(path.Join(f.RunDir, ExperimentResultPath), rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}
