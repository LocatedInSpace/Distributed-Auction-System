// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DASClient is the client API for DAS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DASClient interface {
	Bid(ctx context.Context, in *Amount, opts ...grpc.CallOption) (*Ack, error)
	Result(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Outcome, error)
	// a client can tell the server it has something to sell
	// this is how the active replicas get synced for auctions
	StartAuction(ctx context.Context, in *Item, opts ...grpc.CallOption) (*Ack, error)
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type dASClient struct {
	cc grpc.ClientConnInterface
}

func NewDASClient(cc grpc.ClientConnInterface) DASClient {
	return &dASClient{cc}
}

func (c *dASClient) Bid(ctx context.Context, in *Amount, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, "/proto.DAS/Bid", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dASClient) Result(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Outcome, error) {
	out := new(Outcome)
	err := c.cc.Invoke(ctx, "/proto.DAS/Result", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dASClient) StartAuction(ctx context.Context, in *Item, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, "/proto.DAS/StartAuction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dASClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.DAS/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DASServer is the server API for DAS service.
// All implementations must embed UnimplementedDASServer
// for forward compatibility
type DASServer interface {
	Bid(context.Context, *Amount) (*Ack, error)
	Result(context.Context, *Empty) (*Outcome, error)
	// a client can tell the server it has something to sell
	// this is how the active replicas get synced for auctions
	StartAuction(context.Context, *Item) (*Ack, error)
	Ping(context.Context, *Empty) (*Empty, error)
	mustEmbedUnimplementedDASServer()
}

// UnimplementedDASServer must be embedded to have forward compatible implementations.
type UnimplementedDASServer struct {
}

func (UnimplementedDASServer) Bid(context.Context, *Amount) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bid not implemented")
}
func (UnimplementedDASServer) Result(context.Context, *Empty) (*Outcome, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Result not implemented")
}
func (UnimplementedDASServer) StartAuction(context.Context, *Item) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartAuction not implemented")
}
func (UnimplementedDASServer) Ping(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedDASServer) mustEmbedUnimplementedDASServer() {}

// UnsafeDASServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DASServer will
// result in compilation errors.
type UnsafeDASServer interface {
	mustEmbedUnimplementedDASServer()
}

func RegisterDASServer(s grpc.ServiceRegistrar, srv DASServer) {
	s.RegisterService(&DAS_ServiceDesc, srv)
}

func _DAS_Bid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Amount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DASServer).Bid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DAS/Bid",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DASServer).Bid(ctx, req.(*Amount))
	}
	return interceptor(ctx, in, info, handler)
}

func _DAS_Result_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DASServer).Result(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DAS/Result",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DASServer).Result(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _DAS_StartAuction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Item)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DASServer).StartAuction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DAS/StartAuction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DASServer).StartAuction(ctx, req.(*Item))
	}
	return interceptor(ctx, in, info, handler)
}

func _DAS_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DASServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DAS/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DASServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// DAS_ServiceDesc is the grpc.ServiceDesc for DAS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DAS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.DAS",
	HandlerType: (*DASServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bid",
			Handler:    _DAS_Bid_Handler,
		},
		{
			MethodName: "Result",
			Handler:    _DAS_Result_Handler,
		},
		{
			MethodName: "StartAuction",
			Handler:    _DAS_StartAuction_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _DAS_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/das.proto",
}
