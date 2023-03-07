package service

// gRPC

// get route implementation

/*
Crypto hash user between 0-1000
GetVariant on the data structure
return answer or error
*/

// set route implementation
/*
Do in memory collect & agggregate, with time stamps, with write lock
Periodically flush with read lock to metrics DB
*/

// write metric implementation
/*
Do in memory collect with time stamps with write lock
Periodically flush to metrics DB with read lock
*/

// write blob metrics
/*
Write to metrics DB with write lock...
Interceptors before writing to metricsDB may be useful ...
Consider alongside dashboard interceptors
*/

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/examples/data"

	"github.com/iter8-tools/iter8/controllers"
	sdk "github.com/iter8-tools/iter8/service/sdk"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 50051, "The server port")
)

type sdkServer struct {
	sdk.UnimplementedSDKServer
}

// Return a variant for the given RequestMeta (subject/user ID combination)
//
// If the input is invalid, or if there is no valid variant, then a nil Variant is returned
// along with an error status
func (s *sdkServer) GetRoute(ctx context.Context, req *sdk.RequestMeta) (*sdk.Variant, error) {
	// get variant
	variant, err := controllers.GetVariantFromRequestMeta(req)
	// if none, return error
	if variant == nil {
		return nil, errors.New("no available variant")
	}
	go s.SetControllerVariant(ctx, req.User, variant)
	return variant, nil
}

// Record a route in Iter8 storage
func (s *sdkServer) SetRoute(ctx context.Context, route *sdk.Route) (*emptypb.Empty, error) {
	// get variant
	variant, err := controllers.GetVariantFromNumber(route.Subject, route.Number)
	// if none, return error
	if variant == nil {
		return nil, errors.New("cannot get variant from number")
	}
	go s.SetControllerVariant(ctx, route.User, variant)
	return nil, nil
}

// Record a counter metric value in Iter8 storage
// Metrics recorded by this API correspond to business metrics of an app related to an end-user,
// for example, value of shopping cart purchased by some user of a shopping app in a transaction
func (s *sdkServer) WriteUserMetric(context.Context, *sdk.Counter) (*emptypb.Empty, error) {
	// write to user metrics cache here ...
	// the cache will periodically flush to the database
	return nil, nil
}

// Record a metric blob in Iter8 storage
// Blobs recorded by this API correspond to output of performance testing tools,
// for example, serialized JSON output produced by Fortio or ghz
func (s *sdkServer) WriteMetricBlob(context.Context, *sdk.Blob) (*emptypb.Empty, error) {
	// write to blob metrics cache here ...
	// the cache will periodically flush to the database
	return nil, nil
}

func newServer() *sdkServer {
	return &sdkServer{}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = data.Path("x509/server_cert.pem")
		}
		if *keyFile == "" {
			*keyFile = data.Path("x509/server_key.pem")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	sdk.RegisterSDKServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
