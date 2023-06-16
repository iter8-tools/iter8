package abn

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/controllers/storageclient/badgerdb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		"no application": {namespace: "default", name: "noapp", user: "user", errorSubstring: "routemap not found for application default/noapp"},
		"no user":        {namespace: "default", name: "application", user: "", errorSubstring: "no user session provided"},
		"valid":          {namespace: "default", name: "application", user: "user", errorSubstring: ""},
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
	setupRouteMaps(t, "default", "application")

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
		"no application": {namespace: "", name: "", user: "user", errorSubstring: "routemap not found for application", metric: "", value: "76"},
		"no user":        {namespace: "default", name: "application", user: "", errorSubstring: "no user session provided", metric: "", value: "76"},
		"invalid value":  {namespace: "default", name: "application", user: "user", errorSubstring: "strconv.ParseFloat: parsing \"abc\": invalid syntax", metric: "", value: "abc"},
		"valid":          {namespace: "default", name: "application", user: "user", errorSubstring: "", metric: "metric1", value: "76"},
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

	setupRouteMaps(t, "default", "application")

	if scenario.metric != "" {
		rm, track, err := lookupInternal(scenario.namespace+"/"+scenario.name, scenario.user)
		assert.NoError(t, err)
		assert.NotNil(t, rm)
		assert.NotNil(t, track)

		oldCount = getMetricCountUint64(t, scenario.namespace, scenario.name, *track, scenario.metric)
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
		err := writeMetricInternal(scenario.namespace+"/"+scenario.name, scenario.user, scenario.metric, scenario.value)
		assert.NoError(t, err)
	}

	// verify that metric count has increased by 1
	if scenario.metric != "" {
		rm, track, err := lookupInternal(scenario.namespace+"/"+scenario.name, scenario.user)
		assert.NoError(t, err)
		assert.NotNil(t, rm)
		assert.NotNil(t, track)

		assert.Equal(t, oldCount+1, getMetricCountUint64(t, scenario.namespace, scenario.name, *track, scenario.metric))
	}
}

func TestGetApplicationData(t *testing.T) {
	grpcClient, teardown := setupGRPCService(t)
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := (*grpcClient).GetApplicationData(
		ctx,
		&pb.ApplicationRequest{
			Application: "namespace/doesnotexist",
		},
	)

	assert.Error(t, err)
	assert.Nil(t, s)

}

func setupRouteMaps(t *testing.T, namespace string, name string) {
	controllers.AllRoutemaps.Clear()

	rm := &controllers.Routemap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		// Versions: make([]controllers.Version, 2),
		Versions: []controllers.Version{
			{Signature: util.StringPointer("123456789")},
			{Signature: util.StringPointer("987654321")},
		},
	}

	controllers.AllRoutemaps.AddRouteMap(namespace, name, rm)
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

func getMetricCountUint64(t *testing.T, namespace string, name string, track int, metric string) uint64 {
	rm := controllers.AllRoutemaps.GetRoutemapFromNamespaceName(namespace, name)
	assert.Less(t, track, len(rm.Versions))

	vms, err := metricsClient.GetSummaryMetrics(
		namespace+"/"+name,
		track,
		*rm.Versions[track].Signature,
	)
	assert.NoError(t, err)
	ms := vms.MetricSummaries[metric]

	return ms.SummaryOverTransactions.Count
}

func TestLaunchGRPCServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	metricsPath = t.TempDir()
	err := LaunchGRPCServer(50051, []grpc.ServerOption{}, ctx.Done())
	assert.NoError(t, err)
}
