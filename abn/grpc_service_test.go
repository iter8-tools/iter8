package abn

import (
	"context"
	"net"
	"testing"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"helm.sh/helm/v3/pkg/cli"
)

var testDriver *driver.KubeDriver

type Scenario struct {
	// parameters to lookup
	application string
	user        string
	// expected results for lookup scenarios
	errorSubstring string
	track          string
	// additional expected results for writemetric scenarios
	metric string
	value  string
}

func TestLookup(t *testing.T) {
	testcases := map[string]Scenario{
		"no applicaton": {application: "default/noapp", user: "foobar", errorSubstring: "application not found", track: ""},
		"no user":       {application: "default/application", user: "", errorSubstring: "no user session provided", track: ""},
		"valid":         {application: "default/application", user: "user", errorSubstring: "", track: "candidate"},
	}

	for label, scenario := range testcases {
		t.Run(label, func(t *testing.T) {
			client, teardown := setup(t)
			defer teardown()
			testLookup(t, client, scenario)
		})
	}
}

func testLookup(t *testing.T, client *pb.ABNClient, scenario Scenario) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := (*client).Lookup(
		ctx,
		&pb.Application{
			Name: scenario.application,
			User: scenario.user,
		},
	)

	if scenario.errorSubstring != "" {
		assert.Error(t, err)
		assert.ErrorContains(t, err, scenario.errorSubstring)
	} else {
		assert.NoError(t, err)

		assert.NotNil(t, s)
		assert.Equal(t, scenario.track, s.GetTrack())
	}
}

func TestWriteMetric(t *testing.T) {
	// no application
	// no track from lookup (no user or no tracks)
	// invalid value
	// valid app, track and value

	testcases := map[string]Scenario{
		"no applicaton": {application: "default/noapp", user: "user", errorSubstring: "track not mapped", track: "", metric: "", value: "76"},
		"no user":       {application: "default/application", user: "", errorSubstring: "no user session provided", track: "", metric: "", value: "76"},
		"invalid value": {application: "default/application", user: "user", errorSubstring: "", track: "", metric: "", value: "abc"},
		"valid":         {application: "default/application", user: "user", errorSubstring: "", track: "candidate", metric: "metric1", value: "76"},
	}

	for label, scenario := range testcases {
		t.Run(label, func(t *testing.T) {
			client, teardown := setup(t)
			defer teardown()
			testWriteMetric(t, client, scenario)
		})
	}
}

func testWriteMetric(t *testing.T, client *pb.ABNClient, scenario Scenario) {
	rw := &abnapp.ApplicationReaderWriter{Client: testDriver.Clientset}

	// get current count of metric
	var oldCount uint32 = 0
	var a *abnapp.Application
	a, _ = watcher.GetApplication(scenario.application, rw)
	assert.NotNil(t, a)
	if scenario.metric != "" {
		m := getMetric(a, scenario.track, scenario.metric)
		assert.NotNil(t, m)
		oldCount = m.Count()
	}

	// call gRPC service WriteMetric()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := (*client).WriteMetric(
		ctx,
		&pb.MetricValue{
			Name:        scenario.metric,
			Value:       scenario.value,
			Application: scenario.application,
			User:        scenario.user,
		},
	)

	if scenario.errorSubstring != "" {
		assert.ErrorContains(t, err, scenario.errorSubstring)
	} else {
		assert.NoError(t, err) // never any errors
	}

	// verify that metric count has increased by 1
	a, _ = watcher.GetApplication(scenario.application, rw)
	if scenario.metric != "" {
		m := getMetric(a, scenario.track, scenario.metric)
		assert.NotNil(t, m)
		assert.Equal(t, oldCount+1, m.Count())
	}
}

func setup(t *testing.T) (*pb.ABNClient, func()) {
	testDriver = driver.NewFakeKubeDriver(cli.New())

	// populate watcher.Applications with test applications
	watcher.Applications = map[string]*abnapp.Application{}
	a, err := abnapp.YamlToApplication("default/application", "../../testdata", "abninputs/readtest.yaml")
	a.Writer = &abnapp.ApplicationReaderWriter{Client: testDriver.Clientset}
	assert.NoError(t, err)
	watcher.Applications["default/application"] = a

	// start server
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)

	serverOptions := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterABNServer(grpcServer, newServer(testDriver))
	go grpcServer.Serve(lis)

	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(lis.Addr().String(), opts...)
	assert.NoError(t, err)

	c := pb.NewABNClient(conn)

	// return client and teardown function to clean up
	return &c, func() {
		grpcServer.Stop()
		lis.Close()
		conn.Close()
	}
}

func getMetric(a *abnapp.Application, track, metric string) *abnapp.SummaryMetric {
	version, ok := a.Tracks[track]
	if !ok {
		return nil
	}
	v, _ := a.GetVersion(version, true)
	m, _ := v.GetMetric(metric, true)
	return m
}
