package abn

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/storage/badgerdb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Scenario struct {
	// parameters to lookup
	namespace string
	name      string
	user      string
	// expected results for lookup scenarios
	errorSubstring string
	// additional expected results for writemetric scenarios
	metric string
	value  string
}

func TestLookup(t *testing.T) {
	testcases := map[string]Scenario{
		"no such app": {namespace: "default", name: "noapp", user: "user", errorSubstring: "routemap not found for application default/noapp"},
		"no app":      {namespace: "", name: "", user: "user", errorSubstring: "no application provided"},
		"no user":     {namespace: "default", name: "application", user: "", errorSubstring: "no user session provided"},
		// "valid":       {namespace: "default", name: "application", user: "user", errorSubstring: ""},
	}

	for label, scenario := range testcases {
		t.Run(label, func(t *testing.T) {
			client, teardown := setupGRPCService(t)
			defer teardown()
			testLookup(t, client, scenario)
		})
	}
}
func testLookup(t *testing.T, grpcClient *pb.ABNClient, scenario Scenario) {
	// allroutemaps := setupRoutemaps(t, *getTestRM("default", "application"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := (*grpcClient).Lookup(
		ctx,
		&pb.Application{
			Name: scenario.namespace + "/" + scenario.name,
			User: scenario.user,
		},
	)

	if scenario.errorSubstring != "" {
		assert.Error(t, err)
		assert.ErrorContains(t, err, scenario.errorSubstring)
	} else {
		assert.NoError(t, err)
		assert.NotNil(t, s)
	}
}

func TestWriteMetric(t *testing.T) {
	testcases := map[string]Scenario{
		"no application": {namespace: "", name: "", user: "user", errorSubstring: "no application provided", metric: "", value: "76"},
		"no user":        {namespace: "default", name: "application", user: "", errorSubstring: "no user session provided", metric: "", value: "76"},
		// "invalid value":  {namespace: "default", name: "application", user: "user", errorSubstring: "strconv.ParseFloat: parsing \"abc\": invalid syntax", metric: "", value: "abc"},
		// "valid":          {namespace: "default", name: "application", user: "user", errorSubstring: "", metric: "metric1", value: "76"},
	}

	for label, scenario := range testcases {
		t.Run(label, func(t *testing.T) {
			client, teardown := setupGRPCService(t)
			defer teardown()

			// abnapp.BatchWriteInterval = time.Duration(0)
			testWriteMetric(t, client, scenario)
		})
	}
}

func testWriteMetric(t *testing.T, grpcClient *pb.ABNClient, scenario Scenario) {
	// get current count of metric
	var oldCount uint64

	allroutemaps := setupRoutemaps(t, *getTestRM("default", "application"))

	if scenario.metric != "" {
		rm, track, err := lookupInternal(scenario.namespace+"/"+scenario.name, scenario.user, &allroutemaps)
		assert.NoError(t, err)
		assert.NotNil(t, rm)
		assert.NotNil(t, track)

		oldCount = getMetricCountUint64(t, allroutemaps, scenario.namespace, scenario.name, *track, scenario.metric)
	}

	if scenario.errorSubstring != "" {
		// call gRPC service WriteMetric()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := (*grpcClient).WriteMetric(
			ctx,
			&pb.MetricValue{
				Name:        scenario.metric,
				Value:       scenario.value,
				Application: scenario.namespace + "/" + scenario.name,
				User:        scenario.user,
			},
		)
		assert.ErrorContains(t, err, scenario.errorSubstring)
	} else {
		err := writeMetricInternal(scenario.namespace+"/"+scenario.name, scenario.user, scenario.metric, scenario.value, &allroutemaps)
		assert.NoError(t, err)
	}

	// verify that metric count has increased by 1
	if scenario.metric != "" {
		rm, track, err := lookupInternal(scenario.namespace+"/"+scenario.name, scenario.user, &allroutemaps)
		assert.NoError(t, err)
		assert.NotNil(t, rm)
		assert.NotNil(t, track)

		assert.Equal(t, oldCount+1, getMetricCountUint64(t, allroutemaps, scenario.namespace, scenario.name, *track, scenario.metric))
	}
}

func setupRoutemaps(t *testing.T, initialroutemaps ...testroutemap) testroutemaps {
	routemaps := testroutemaps{
		nsRoutemap: make(map[string]testroutemapsByName),
	}

	for i := range initialroutemaps {

		if _, ok := routemaps.nsRoutemap[initialroutemaps[i].namespace]; !ok {
			routemaps.nsRoutemap[initialroutemaps[i].namespace] = make(testroutemapsByName)
		}
		(routemaps.nsRoutemap[initialroutemaps[i].namespace])[initialroutemaps[i].name] = &initialroutemaps[i]
	}

	return routemaps
}

func getTestRM(namespace, name string) *testroutemap {
	return &testroutemap{
		namespace: namespace,
		name:      name,
		versions: []testversion{
			{signature: util.StringPointer("123456789")},
			{signature: util.StringPointer("987654321")},
		},
		normalizedWeights: []uint32{1, 1},
	}

}

func setupGRPCService(t *testing.T) (*pb.ABNClient, func()) {
	// 49152-65535 are recommended ports; we use a random one for testing
	/* #nosec */
	port := rand.Intn(65535-49152) + 49152

	// start server
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	assert.NoError(t, err)

	serverOptions := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterABNServer(grpcServer, newServer())
	tempDirPath := t.TempDir()
	metricsClient, err = badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)
	go func() {
		_ = grpcServer.Serve(lis)
	}()

	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(lis.Addr().String(), opts...)
	assert.NoError(t, err)

	c := pb.NewABNClient(conn)

	// return client and teardown function to clean up
	return &c, func() {
		grpcServer.Stop()
		_ = lis.Close()
		_ = conn.Close()
	}

}

func getMetricCountUint64(t *testing.T, routemaps testroutemaps, namespace string, name string, track int, metric string) uint64 {
	rm := routemaps.GetRoutemapFromNamespaceName(namespace, name)
	assert.Less(t, track, len(rm.GetVersions()))

	vms, err := metricsClient.GetSummaryMetrics(
		namespace+"/"+name,
		track,
		*rm.GetVersions()[track].GetSignature(),
	)
	assert.NoError(t, err)
	ms := vms.MetricSummaries[metric]

	return ms.SummaryOverTransactions.Count
}

func TestLaunchGRPCServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	metricsPath = t.TempDir()

	configFile := filepath.Clean(util.CompletePath("../testdata", "abninputs/config.yaml"))
	err := os.Setenv("CONFIG_FILE", configFile)
	assert.NoError(t, err)

	err = LaunchGRPCServer(50051, []grpc.ServerOption{}, ctx.Done())
	assert.NoError(t, err)
}
