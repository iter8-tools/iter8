package abn

// service.go - entry point for A/B/n service

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/iter8-tools/iter8/abn/grpc"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	"helm.sh/helm/v3/pkg/cli"

	// auth package is necessary to enable authentication with various cloud providers
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
	)

	if err != nil {
		log.Logger.Warn("Lookup failed: ", err)
		return nil, err
	}

	if track == nil {
		log.Logger.Warn("lookup returned nil")
		return nil, err
	}

	log.Logger.Debugf("lookup(%s,%s) -> %d", appMsg.GetName(), appMsg.GetUser(), *track)

	return &pb.Session{
		Track: fmt.Sprintf("%d", *track),
	}, err
}

// WriteMetric identifies the track with which a metric is associated (from user) and
// writes the metric value (currently only supports summary metrics)
func (server *abnServer) WriteMetric(ctx context.Context, metricMsg *pb.MetricValue) (*emptypb.Empty, error) {
	log.Logger.Trace("WriteMetric called")
	defer log.Logger.Trace("WriteMetric completed")

	client, err := k8sclient.New(cli.New())
	if err != nil {
		log.Logger.Error("could not obtain Kubernetes client")
		return &emptypb.Empty{}, err
	}

	err = writeMetricInternal(
		metricMsg.GetApplication(),
		metricMsg.GetUser(),
		metricMsg.GetName(),
		metricMsg.GetValue(),
		client,
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

// LaunchGRPCServer starts gRPC server
func LaunchGRPCServer(port int, opts []grpc.ServerOption, stopCh <-chan struct{}) error {
	log.Logger.Tracef("starting gRPC service on port %d", port)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Logger.WithError(err).Error("service failed to listen")
		return err
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterABNServer(grpcServer, newServer())

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
