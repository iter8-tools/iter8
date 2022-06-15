package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/abn/watcher"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Currently, the track is not updated if a second object with a different track is identified
// Is this the right approach?  Should it be updated to the latest track?
// Should an error be registered?
// Currently, if the version (or track) is modified, the old value is not removed.
// In particular, a version will remain listed even if it is no longer relevant
// this means data loss

const (
	// Name of environment variable with file path to resources/namespaces to watch
	WATCHER_CONFIG_ENV = "WATCHER_CONFIG"
)

var (
	// Port the service listens on
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	log.Logger.SetLevel(logrus.TraceLevel)

	// read abn config (resources and namespaces to watch)
	abnConfigFile, ok := os.LookupEnv(WATCHER_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("ABn configuation file is required")
	}

	stopCh := make(chan struct{})

	// set up watching
	go newInformer(watcher.ReadConfig(abnConfigFile)).Start(stopCh)

	// launch gRPC server to respond to frontend requests
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

// newInformer creates a new informer watching the identified resources in the identified namespaces
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

// newServer returns a new gRPC server
func newServer() *abnServer {
	s := &abnServer{}
	return s
}

type abnServer struct {
	pb.UnimplementedABNServer
}

// Lookup identifies a version that should be used for a given user
// This method is exposed to gRPC clients
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
	}, err
}

type MetricEntry struct {
	name        string
	value       string
	application string
	user        string
	track       string
	version     string
	time        string
}

// WriteMetric writes a metric
// This implmementation writes the metric to the log
// This method is exposed to gRPC clients
func (server *abnServer) WriteMetric(ctx context.Context, m *pb.MetricValue) (*emptypb.Empty, error) {
	v, err := watcher.Lookup(m.GetApplication(), m.GetUser())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	track := v.Track
	if track == "" {
		track = v.Name
	}

	me := MetricEntry{
		name:        m.GetName(),
		value:       m.GetValue(),
		application: m.GetApplication(),
		user:        m.GetUser(),
		track:       track,
		version:     v.Name,
		time:        time.Now().UTC().Format("2006-01-02 15:04:05"),
	}

	log.Logger.Info("WriteMetric: ", me)
	return &emptypb.Empty{}, nil
}

// launchServer starts gRPC server
func launchServer(opts []grpc.ServerOption) {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
