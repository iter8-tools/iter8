package abn

// service.go - entry point for A/B(/n) service; starts controller watching resources
//               and gRPC service to respond to lookup and write metric requests

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"

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
func Start(kd *base.KubeDriver) {
	// initialize kubernetes driver
	if err := kd.InitKube(); err != nil {
		log.Logger.Fatal("unable to initialize kubedriver")
	}

	// read abn config (resources and namespaces to watch)
	abnConfigFile, ok := os.LookupEnv(WATCHER_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("configuation file is required")
	}

	stopCh := make(chan struct{})

	// set up resource watching as defined by config
	abnConfig := watcher.ReadConfig(abnConfigFile)
	w := watcher.NewIter8Watcher(kd, abnConfig.Resources, abnConfig.Namespaces)
	go w.Start(stopCh)

	// launch gRPC server to respond to frontend requests
	go launchGRPCServer([]grpc.ServerOption{}, kd)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	<-sigCh
	close(stopCh)
}

// newServer returns a new gRPC server
func newServer(kd *base.KubeDriver) *abnServer {
	s := &abnServer{
		Driver: kd,
	}
	return s
}

type abnServer struct {
	Driver *base.KubeDriver
	pb.UnimplementedABNServer
}

// Lookup identifies a version that should be used for a given user
// This method is exposed to gRPC clients
func (server *abnServer) Lookup(ctx context.Context, appMsg *pb.Application) (*pb.Session, error) {
	watcher.Applications.Lock()
	defer watcher.Applications.Unlock()

	track, err := pb.Lookup(appMsg.GetName(), appMsg.GetUser())
	if err != nil || track == nil {
		return nil, err
	}

	return &pb.Session{
		Track: *track,
	}, err
}

func (server *abnServer) WriteMetric(ctx context.Context, metricMsg *pb.MetricValue) (*emptypb.Empty, error) {
	watcher.Applications.Lock()
	defer watcher.Applications.Unlock()

	a, err := watcher.Applications.Get(metricMsg.Application, nil)
	// a, ok := watcher.Applications[metricMsg.GetApplication()]
	if err != nil || a == nil {
		// if !ok {
		return &emptypb.Empty{}, errors.New("unexpected: cannot find record of application " + metricMsg.GetApplication())
	}

	track, err := pb.Lookup(metricMsg.GetApplication(), metricMsg.GetUser())
	if err != nil || track == nil {
		return &emptypb.Empty{}, err
	}

	version, ok := a.Tracks[*track]
	if !ok {
		return &emptypb.Empty{}, errors.New("track not mapped to version")
	}

	v, _ := a.GetVersion(version, false)
	if v == nil {
		return &emptypb.Empty{}, errors.New("unexpected: trying to write metrics for unknown version")
	}

	value, err := strconv.ParseFloat(metricMsg.GetValue(), 64)
	if err != nil {
		log.Logger.Warn("Unable to parse metric value ", metricMsg.GetValue())
		return &emptypb.Empty{}, nil
	}

	m, _ := v.GetMetric(metricMsg.GetName(), true)
	m.Add(value)

	// persist updated metric
	a.Write()

	return &emptypb.Empty{}, nil
}

// launchGRPCServer starts gRPC server
func launchGRPCServer(opts []grpc.ServerOption, kd *base.KubeDriver) {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer(kd))
	grpcServer.Serve(lis)
}
