package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base/log"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	WATCHER_CONFIG_ENV = "WATCHER_CONFIG"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()

	// abn config
	abnConfigFile, ok := os.LookupEnv(WATCHER_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("ABn configuation file is required")
	}

	stopCh := make(chan struct{})
	go newInformer(watcher.ReadConfig(abnConfigFile)).Start(stopCh)

	go launchServer([]grpc.ServerOption{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Kill, os.Interrupt)

	<-sigCh
	close(stopCh)
}

func restConfig() (*rest.Config, error) {
	kubeCfg, err := rest.InClusterConfig()
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		kubeCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	return kubeCfg, nil
}

func newInformer(abnConfig *watcher.Config) *watcher.MultiInformer {
	cfg, err := restConfig()
	if err != nil {
		log.Logger.WithError(err).Fatal("could not get REST config")
	}

	client, err := watcher.NewClient(cfg)
	if err != nil {
		log.Logger.WithError(err).Fatal("unable to create watcher client")
	}

	return watcher.NewInformer(
		client,
		abnConfig.Resources,
		abnConfig.Namespaces,
	)
}

func newServer() *abnServer {
	s := &abnServer{}
	return s
}

type abnServer struct {
	pb.UnimplementedABNServer
}

func (server *abnServer) Lookup(ctx context.Context, a *pb.Application) (*pb.Session, error) {
	v, err := watcher.Lookup(a.GetName(), a.GetUser())
	if err != nil {
		return nil, err
	}
	track := v.Track
	if track == "" {
		track = v.Name
	}
	return &pb.Session{
		Track: track,
	}, nil
}

func (server *abnServer) WriteMetric(ctx context.Context, m *pb.MetricValue) (*emptypb.Empty, error) {
	return nil, nil
}

func launchServer(opts []grpc.ServerOption) {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
