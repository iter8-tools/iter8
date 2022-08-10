// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ABNClient is the client API for ABN service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ABNClient interface {
	// Identify a track the caller should send a request to.
	// Should be called for each request (transaction).
	Lookup(ctx context.Context, in *Application, opts ...grpc.CallOption) (*Session, error)
	// Write a metric value to metrics database.
	// The metric value is explicitly associated with a list of transactions that contributed to its computation.
	// The user is expected to identify these transactions.
	WriteMetric(ctx context.Context, in *MetricValue, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type aBNClient struct {
	cc grpc.ClientConnInterface
}

func NewABNClient(cc grpc.ClientConnInterface) ABNClient {
	return &aBNClient{cc}
}

func (c *aBNClient) Lookup(ctx context.Context, in *Application, opts ...grpc.CallOption) (*Session, error) {
	out := new(Session)
	err := c.cc.Invoke(ctx, "/main.ABN/Lookup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aBNClient) WriteMetric(ctx context.Context, in *MetricValue, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/main.ABN/WriteMetric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ABNServer is the server API for ABN service.
// All implementations must embed UnimplementedABNServer
// for forward compatibility
type ABNServer interface {
	// Identify a track the caller should send a request to.
	// Should be called for each request (transaction).
	Lookup(context.Context, *Application) (*Session, error)
	// Write a metric value to metrics database.
	// The metric value is explicitly associated with a list of transactions that contributed to its computation.
	// The user is expected to identify these transactions.
	WriteMetric(context.Context, *MetricValue) (*emptypb.Empty, error)
	mustEmbedUnimplementedABNServer()
}

// UnimplementedABNServer must be embedded to have forward compatible implementations.
type UnimplementedABNServer struct {
}

func (UnimplementedABNServer) Lookup(context.Context, *Application) (*Session, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lookup not implemented")
}
func (UnimplementedABNServer) WriteMetric(context.Context, *MetricValue) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WriteMetric not implemented")
}
func (UnimplementedABNServer) mustEmbedUnimplementedABNServer() {}

// UnsafeABNServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ABNServer will
// result in compilation errors.
type UnsafeABNServer interface {
	mustEmbedUnimplementedABNServer()
}

func RegisterABNServer(s grpc.ServiceRegistrar, srv ABNServer) {
	s.RegisterService(&ABN_ServiceDesc, srv)
}

func _ABN_Lookup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Application)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ABNServer).Lookup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.ABN/Lookup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ABNServer).Lookup(ctx, req.(*Application))
	}
	return interceptor(ctx, in, info, handler)
}

func _ABN_WriteMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ABNServer).WriteMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.ABN/WriteMetric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ABNServer).WriteMetric(ctx, req.(*MetricValue))
	}
	return interceptor(ctx, in, info, handler)
}

// ABN_ServiceDesc is the grpc.ServiceDesc for ABN service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ABN_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "main.ABN",
	HandlerType: (*ABNServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Lookup",
			Handler:    _ABN_Lookup_Handler,
		},
		{
			MethodName: "WriteMetric",
			Handler:    _ABN_WriteMetric_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "abn/grpc/abn.proto",
}
