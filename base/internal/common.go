package internal

/*
Credits: This file is sourced from https://github.com/bojand/ghz and modified for reuse in Iter8
*/

import (
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/iter8-tools/iter8/base/internal/helloworld/helloworld"
)

// TestPort is the port.
var TestPort string

// TestLocalhost is the localhost.
var TestLocalhost string

// StartServer starts the server.
//
// For testing only.
func StartServer(secure bool) (*helloworld.Greeter, *grpc.Server, error) {
	lis, err := net.Listen("tcp", ":0")
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

	TestPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	TestLocalhost = "localhost:" + TestPort

	go func() {
		_ = s.Serve(lis)
	}()

	return gs, s, err
}
