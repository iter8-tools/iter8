package action

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	// fix gOpts
	os.Chdir(t.TempDir())
	gOpts := NewGenOpts()
	gOpts.ChartsParentDir = base.CompletePath("../", "")
	gOpts.ChartName = "load-test-http"
	gOpts.Values = []string{"url=https://httpbin.org/get"}
	err := gOpts.LocalRun()
	assert.NoError(t, err)
}

func TestGenGRPC(t *testing.T) {
	// fix gOpts
	os.Chdir(t.TempDir())
	gOpts := NewGenOpts()
	gOpts.ChartsParentDir = base.CompletePath("../", "")
	gOpts.ChartName = "load-test-grpc"
	gOpts.Values = []string{"host=localhost:50051", "call=helloworld.Greeter.SayHello", "proto=helloworld.proto", "protoset=helloworld.protoset", "data.name=frodo", "SLOs.grpc/error-rate=0", "SLOs.grpc/latency/mean=150"}
	err := gOpts.LocalRun()
	assert.NoError(t, err)

	fd := &driver.FileDriver{
		RunDir: "./",
	}
	exp, err := base.BuildExperiment(false, fd)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(exp.Tasks))

	m := make(map[string]interface{}, 0)
	b, _ := json.Marshal(exp.Tasks[0])
	json.Unmarshal(b, &m)
	m = m["with"].(map[string]interface{})
	s := m["proto"].(string)
	assert.Equal(t, "helloworld.proto", s)

}
