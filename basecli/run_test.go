package basecli

import (
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

type MockIO struct {
	es base.ExperimentSpec
	er *base.ExperimentResult
}

func (n *MockIO) ReadResult() (*base.ExperimentResult, error) {
	return n.er, nil
}

func (n *MockIO) WriteResult(r *Experiment) error {
	return nil
}

func (n *MockIO) ReadSpec() (base.ExperimentSpec, error) {
	return n.es, nil
}

func TestRun(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.Deactivate)
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `all good`))

	b, err := ioutil.ReadFile(base.CompletePath("../testdata", "experiment.yaml"))
	assert.NoError(t, err)
	es := &base.ExperimentSpec{}
	err = yaml.Unmarshal(b, es)
	assert.NoError(t, err)
	exp := Experiment{
		Experiment: base.Experiment{
			Tasks: *es,
		},
	}

	err = exp.Run(&MockIO{})
	assert.NoError(t, err)
}
