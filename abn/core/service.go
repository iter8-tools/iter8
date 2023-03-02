package core

// service.go - entry point for A/B/n service; starts controller watching resources
//               and gRPC service to respond to lookup and write metric requests

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base/log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	// auth package is necessary to enable authentication with various cloud vendors
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to configuration yaml file
	watcherConfigEnv = "WATCHER_CONFIG"
)

// Start is entry point to configure services and start them
func Start(port int, stopCh chan struct{}) error {
	if err := k8sclient.Client.Initialize(); err != nil {
		log.Logger.WithStackTrace("unable to initialize k8s client").Fatal(err)
	}

	// read abn config (resources and namespaces to watch)
	abnConfigFile, ok := os.LookupEnv(watcherConfigEnv)
	if !ok {
		log.Logger.Fatal("configuation file is required")
	}

	// set up resource watching as defined by config
	w := watcher.NewIter8Watcher(abnConfigFile)

	// start watchers
	go w.Start(stopCh)

	// launch gRPC server to respond to frontend requests
	go launchGRPCServer(port, []grpc.ServerOption{}, stopCh)

	return nil
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
	log.Logger.Debug(fmt.Sprintf("lookup(%s,%s) -> %s", appMsg.GetName(), appMsg.GetUser(), *track))

	return &pb.Session{
		Track: *track,
	}, err
}

// WriteMetric identifies the track with which a metric is associated (from user) and
// writes the metric value (currently only supports summary metrics)
func (server *abnServer) WriteMetric(ctx context.Context, metricMsg *pb.MetricValue) (*emptypb.Empty, error) {
	log.Logger.Trace("WriteMetric called")
	defer log.Logger.Trace("WriteMetric completed")

	err := writeMetricInternal(
		metricMsg.GetApplication(),
		metricMsg.GetUser(),
		metricMsg.GetName(),
		metricMsg.GetValue(),
	)

	return &emptypb.Empty{}, err
}

func (server *abnServer) GetApplicationData(ctx context.Context, metricReqMsg *pb.ApplicationRequest) (*pb.ApplicationData, error) {
	jsonStr, err := getApplicationDataInternal(
		metricReqMsg.GetApplication(),
	)

	return &pb.ApplicationData{
		ApplicationJson: jsonStr,
	}, err
}

// launchGRPCServer starts gRPC server
func launchGRPCServer(port int, opts []grpc.ServerOption, stopCh chan struct{}) {
	log.Logger.Tracef("starting gRPC service on port %d", port)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-stopCh
		log.Logger.Warnf("stop channel closed, shutting down")
		grpcServer.GracefulStop()
	}()

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to start service")
	}
	wg.Wait()
	log.Logger.Trace("service shutdown")
}
