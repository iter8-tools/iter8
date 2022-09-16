package internal

/*
Credit: This file is sourced from https://github.com/bojand/ghz and modified for reuse in Iter8
*/

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/iter8-tools/iter8/base/internal/helloworld/helloworld"
)

// LocalHostPort is the localhost:12345 combo used for testing
const LocalHostPort = "localhost:12345"

// StartServer starts the server.
//
// For testing only.
func StartServer(secure bool) (*helloworld.Greeter, *grpc.Server, error) {
	lis, err := net.Listen("tcp", LocalHostPort)
	if err != nil {
		return nil, nil, err
	}

	var opts []grpc.ServerOption

	stats := helloworld.NewHWStats()

	opts = append(opts, grpc.StatsHandler(stats))

	s := grpc.NewServer(opts...)

	gs := helloworld.NewGreeter()
	helloworld.RegisterGreeterServer(s, gs)
	reflection.Register(s)

	gs.Stats = stats

	go func() {
		_ = s.Serve(lis)
	}()

	return gs, s, err
}
