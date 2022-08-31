package abn

// service.go - entry point for A/B(/n) service; starts controller watching resources
//               and gRPC service to respond to lookup and write metric requests

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base/log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to configuration yaml file
	WATCHER_CONFIG_ENV = "WATCHER_CONFIG"
)

var (
	// port the service listens on
	port = flag.Int("port", 50051, "The server port")
)

// Start is entry point to configure services and start them
func Start(kClient *k8sclient.KubeClient) {
	// Initialize kubernetes client
	if err := kClient.Initialize(); err != nil {
		log.Logger.Fatal("unable to initialize kubernetes client")
	}

	// read abn config (resources and namespaces to watch)
	abnConfigFile, ok := os.LookupEnv(WATCHER_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("configuation file is required")
	}

	stopCh := make(chan struct{})

	// set up resource watching as defined by config
	c := readConfig(abnConfigFile)
	w := watcher.NewIter8Watcher(kClient, c.Resources, c.Namespaces)
	go w.Start(stopCh)

	// launch gRPC server to respond to frontend requests
	go launchGRPCServer([]grpc.ServerOption{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	<-sigCh
	close(stopCh)
}

// newServer returns a new gRPC server
func newServer() *abnServer {
	return &abnServer{}
}

type abnServer struct {
	pb.UnimplementedABNServer
}

// Lookup identifies a track that should be used for a given user
// This method is exposed to gRPC clients
func (server *abnServer) Lookup(ctx context.Context, appMsg *pb.Application) (*pb.Session, error) {
	_, track, err := lookupInternal(
		appMsg.GetName(),
		appMsg.GetUser(),
	)

	if err != nil || track == nil {
		return nil, err
	}
	return &pb.Session{
		Track: *track,
	}, err
}

// WriteMetric identifies the track with which a metric is associated (from user) and
// writes the metric value (currently only supports summary metrics)
func (server *abnServer) WriteMetric(ctx context.Context, metricMsg *pb.MetricValue) (*emptypb.Empty, error) {
	err := writeMetricInternal(
		metricMsg.GetApplication(),
		metricMsg.GetUser(),
		metricMsg.GetName(),
		metricMsg.GetValue(),
	)

	return &emptypb.Empty{}, err
}

// launchGRPCServer starts gRPC server
func launchGRPCServer(opts []grpc.ServerOption) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
