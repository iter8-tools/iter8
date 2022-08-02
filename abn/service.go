package abn

// service.go - entry point for A/B(/n) service; starts controller watching resources and gRPC service

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/metricstore"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

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
func Start(kd *driver.KubeDriver) {
	// initialize kubernetes driver
	if err := kd.Init(); err != nil {
		log.Logger.Fatal("unable to initialize kubedriver")
	}

	// read abn config (resources and namespaces to watch)
	abnConfigFile, ok := os.LookupEnv(WATCHER_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("configuation file is required")
	}

	stopCh := make(chan struct{})

	// set up resource watching as defined by config
	// go newInformer(watcher.ReadConfig(abnConfigFile), namespace+"/"+name).Start(stopCh)
	go newInformer(watcher.ReadConfig(abnConfigFile), kd).Start(stopCh)

	// launch gRPC server to respond to frontend requests
	go launchGRPCServer([]grpc.ServerOption{}, kd)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	<-sigCh
	close(stopCh)
}

// newInformer creates a new informer watching the identified resources in the identified namespaces
// func newInformer(abnConfig watcher.Config, name string) *watcher.MultiInformer {
func newInformer(abnConfig watcher.Config, kd *driver.KubeDriver) *watcher.MultiInformer {
	return watcher.NewInformer(
		kd,
		abnConfig.Resources,
		abnConfig.Namespaces,
	)
}

// newServer returns a new gRPC server
func newServer(kd *driver.KubeDriver) *abnServer {
	s := &abnServer{
		Driver: kd,
	}
	return s
}

type abnServer struct {
	Driver *driver.KubeDriver
	pb.UnimplementedABNServer
}

// Lookup identifies a version that should be used for a given user
// This method is exposed to gRPC clients
func (server *abnServer) Lookup(ctx context.Context, a *pb.Application) (*pb.Session, error) {
	v, err := pb.Lookup(a.GetName(), a.GetUser())
	if err != nil {
		return nil, err
	}
	track := v.Track
	if track == "" {
		track = v.Name
	}
	return &pb.Session{
		Track: track,
	}, err
}

// WriteMetric writes a metric
// This implmementation writes the metric to a Kubernetes secret
// This method is exposed to gRPC clients
func (server *abnServer) WriteMetric(ctx context.Context, m *pb.MetricValue) (*emptypb.Empty, error) {
	v, err := pb.Lookup(m.GetApplication(), m.GetUser())
	if err != nil {
		return &emptypb.Empty{}, err
	}

	ms, err := metricstore.NewMetricStoreSecret(m.GetApplication(), server.Driver)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	value, err := strconv.ParseFloat(m.GetValue(), 64)
	if err != nil {
		log.Logger.Warn("Unable to parse metric value ", m.GetValue())
		return &emptypb.Empty{}, nil
	}
	err = ms.AddMetric(m.GetName(), v.Name, value)
	if err != nil {
		log.Logger.Warn("unable to write metric to metric store")
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, nil
}

// launchGRPCServer starts gRPC server
func launchGRPCServer(opts []grpc.ServerOption, kd *driver.KubeDriver) {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer(kd))
	grpcServer.Serve(lis)
}
