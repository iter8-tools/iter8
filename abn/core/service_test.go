package core

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"helm.sh/helm/v3/pkg/cli"
)

func TestStart(t *testing.T) {
	abnapp.Applications.Clear()
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	// set watcherConfigEnv to test config file
	_ = os.Setenv(watcherConfigEnv, "../../testdata/abninputs/config.yaml")

	stopCh := make(chan struct{})
	defer close(stopCh)

	// 49152-65535 are recommended ports; we use a random one for testing
	/* #nosec */
	port := rand.Intn(65535-49152) + 49152

	err := Start(port, stopCh)
	assert.NoError(t, err)

	// verify grpc service working by calling a method
	// there is no data so should be told not found
	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", port), opts...)
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()
	c := pb.NewABNClient(conn)
	client := &c

	// initially the service might take time to come up
	assert.Eventually(
		t,
		func() bool {
			_, err = callLookup(client)
			return assert.ErrorContains(t, err, "application not found")
		},
		10*time.Second,
		time.Second,
	)
}

func callLookup(client *pb.ABNClient) (*pb.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a, err := (*client).Lookup(
		ctx,
		&pb.Application{
			Name: "default/application",
			User: "user",
		},
	)
	return a, err
}
