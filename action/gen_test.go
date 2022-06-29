package action

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	// fix gOpts
	os.Chdir(t.TempDir())
	gOpts := NewGenOpts()
	gOpts.ChartsParentDir = base.CompletePath("../", "")
	gOpts.ChartName = "iter8"
	gOpts.Values = []string{"tasks={http}", "http.url=https://httpbin.org/get"}
	err := gOpts.LocalRun()
	assert.NoError(t, err)
}

func dumpExperiment(t *testing.T) {
	file, err := os.Open("experiment.yaml")
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)
	file.Close()
	l := log.Logger.GetLevel()
	log.Logger.SetLevel(logrus.DebugLevel)
	log.Logger.Debug("\n" + string(b))
	log.Logger.SetLevel(l)
}

func TestGenGRPC(t *testing.T) {
	// fix gOpts
	os.Chdir(t.TempDir())
	gOpts := NewGenOpts()
	gOpts.ChartsParentDir = base.CompletePath("../", "")
	gOpts.ChartName = "iter8"
	gOpts.Values = []string{"tasks={grpc,assess}", "grpc.host=localhost:50051", "grpc.call=helloworld.Greeter.SayHello", "grpc.proto=helloworld.proto", "grpc.protoset=helloworld.protoset", "grpc.data.name=frodo", "assess.SLOs.upper.grpc/error-rate=0", "assess.SLOs.upper.grpc/latency/mean=150"}
	err := gOpts.LocalRun()
	assert.NoError(t, err)

	fd := &driver.FileDriver{
		RunDir: "./",
	}
	exp, err := base.BuildExperiment(fd)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(exp.Spec))

	m := make(map[string]interface{}, 0)
	b, _ := json.Marshal(exp.Spec[0])
	json.Unmarshal(b, &m)
	m = m["with"].(map[string]interface{})
	s := m["proto"].(string)
	assert.Equal(t, "helloworld.proto", s)
}

func TestGenDB(t *testing.T) {
	// fix gOpts
	os.Chdir(t.TempDir())
	gOpts := NewGenOpts()
	gOpts.ChartsParentDir = base.CompletePath("../", "")
	gOpts.ChartName = "iter8"
	gOpts.Values = []string{"tasks={custommetrics,assess}", "custommetrics.templates.istio-prom=https://raw.githubusercontent.com/iter8-tools/iter8/master/charts/iter8lib/templates/_metrics-istio.tpl", "custommetrics.values.URL=http://prometheus.istio-system:9090/api/v1/query", "custommetrics.values.destinationWorkload=httpbin-v2", "custommetrics.values.destinationWorkloadNamespace=default", `custommetrics.values.startingTime="2020-02-01T09:44:40Z"`, "assess.SLOs.upper.istio/error-rate=0"}

	err := gOpts.LocalRun()
	assert.NoError(t, err)

	dumpExperiment(t)

	fd := &driver.FileDriver{
		RunDir: "./",
	}
	exp, err := base.BuildExperiment(fd)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(exp.Spec))

	m := make(map[string]interface{}, 0)
	b, _ := json.Marshal(exp.Spec[0])
	json.Unmarshal(b, &m)
	m = m["with"].(map[string]interface{})
	s := m["providerURLs"].([]interface{})
	assert.Equal(t, []interface{}{"https://raw.githubusercontent.com/iter8-tools/iter8/master/charts/iter8lib/templates/_metrics-istio.tpl"}, s)
}
