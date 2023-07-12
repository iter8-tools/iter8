// Package abn contains the implementation of the A/B/n service (via gRPC)
package abn

// service.go - entry point for A/B/n service

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/dgraph-io/badger/v4"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/storage"
	"github.com/iter8-tools/iter8/storage/badgerdb"

	// auth package is necessary to enable authentication with various cloud providers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// defaultMetricsPath is the default path of the persistent volume
	defaultMetricsPath = "/metrics"
)

var (
	// MetricsPath is the path of the persistent volume
	metricsPath = defaultMetricsPath
	// MetricsClient is the metrics client
	MetricsClient storage.Interface
)

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
	log.Logger.Tracef("Lookup called for application=%s, user=%s", appMsg.GetName(), appMsg.GetUser())
	defer log.Logger.Trace("Lookup completed")

	_, track, err := lookupInternal(
		appMsg.GetName(),
		appMsg.GetUser(),
		&controllers.AllRoutemaps,
	)

	if err != nil {
		log.Logger.Warnf("Lookup(%s,%s) failed: %s", appMsg.GetName(), appMsg.GetUser(), err.Error())
		return nil, err
	}

	if track == nil {
		log.Logger.Warnf("Lookup(%s,%s) returned nil", appMsg.GetName(), appMsg.GetUser())
		return nil, err
	}

	log.Logger.Tracef("Lookup(%s,%s) -> %d", appMsg.GetName(), appMsg.GetUser(), *track)

	return &pb.Session{
		Track: fmt.Sprintf("%d", *track),
	}, err
}

// WriteMetric identifies the track with which a metric is associated (from user) and
// writes the metric value (currently only supports summary metrics)
func (server *abnServer) WriteMetric(ctx context.Context, metricMsg *pb.MetricValue) (*emptypb.Empty, error) {
	log.Logger.Trace("WriteMetric called")
	defer log.Logger.Trace("WriteMetric completed")

	return &emptypb.Empty{},
		writeMetricInternal(
			metricMsg.GetApplication(),
			metricMsg.GetUser(),
			metricMsg.GetName(),
			metricMsg.GetValue(),
			&controllers.AllRoutemaps,
		)
}

const (
	configEnv         = "ABN_CONFIG_FILE"
	defaultPortNumber = 50051
)

// abnConfig defines the configuration of the controllers
type abnConfig struct {
	// Port is port number on which the abn gRPC service should listen
	Port *int `json:"port,omitempty"`
}

// LaunchGRPCServer starts gRPC server
func LaunchGRPCServer(opts []grpc.ServerOption, stopCh <-chan struct{}) error {
	// read configutation for metrics service
	conf := &abnConfig{}
	err := util.ReadConfig(configEnv, conf, func() {
		if nil == conf.Port {
			conf.Port = util.IntPointer(defaultPortNumber)
		}
	})
	if err != nil {
		log.Logger.Errorf("unable to read metrics configuration: %s", err.Error())
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *conf.Port))
	if err != nil {
		log.Logger.WithError(err).Error("service failed to listen")
		return err
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer())

	// configure metricsClient if needed

	MetricsClient, err = badgerdb.GetClient(badger.DefaultOptions(metricsPath), badgerdb.AdditionalOptions{})
	if err != nil {
		log.Logger.Error("Unable to configure metrics storage client ", err)
		return err
	}

	go func() {
		<-stopCh
		log.Logger.Warnf("stop channel closed, shutting down")
		grpcServer.GracefulStop()
	}()

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Logger.WithError(err).Error("failed to start service")
		return err
	}

	return nil
}
