// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: appsender/appsender.proto

package appsender

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

const (
	AppSender_SendAppRequest_FullMethodName            = "/appsender.AppSender/SendAppRequest"
	AppSender_SendAppResponse_FullMethodName           = "/appsender.AppSender/SendAppResponse"
	AppSender_SendAppError_FullMethodName              = "/appsender.AppSender/SendAppError"
	AppSender_SendAppGossip_FullMethodName             = "/appsender.AppSender/SendAppGossip"
	AppSender_SendCrossChainAppRequest_FullMethodName  = "/appsender.AppSender/SendCrossChainAppRequest"
	AppSender_SendCrossChainAppResponse_FullMethodName = "/appsender.AppSender/SendCrossChainAppResponse"
	AppSender_SendCrossChainAppError_FullMethodName    = "/appsender.AppSender/SendCrossChainAppError"
)

// AppSenderClient is the client API for AppSender service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AppSenderClient interface {
	SendAppRequest(ctx context.Context, in *SendAppRequestMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendAppResponse(ctx context.Context, in *SendAppResponseMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendAppError(ctx context.Context, in *SendAppErrorMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendAppGossip(ctx context.Context, in *SendAppGossipMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendCrossChainAppRequest(ctx context.Context, in *SendCrossChainAppRequestMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendCrossChainAppResponse(ctx context.Context, in *SendCrossChainAppResponseMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendCrossChainAppError(ctx context.Context, in *SendCrossChainAppErrorMsg, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type appSenderClient struct {
	cc grpc.ClientConnInterface
}

func NewAppSenderClient(cc grpc.ClientConnInterface) AppSenderClient {
	return &appSenderClient{cc}
}

func (c *appSenderClient) SendAppRequest(ctx context.Context, in *SendAppRequestMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendAppRequest_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appSenderClient) SendAppResponse(ctx context.Context, in *SendAppResponseMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendAppResponse_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appSenderClient) SendAppError(ctx context.Context, in *SendAppErrorMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendAppError_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appSenderClient) SendAppGossip(ctx context.Context, in *SendAppGossipMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendAppGossip_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appSenderClient) SendCrossChainAppRequest(ctx context.Context, in *SendCrossChainAppRequestMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendCrossChainAppRequest_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appSenderClient) SendCrossChainAppResponse(ctx context.Context, in *SendCrossChainAppResponseMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendCrossChainAppResponse_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appSenderClient) SendCrossChainAppError(ctx context.Context, in *SendCrossChainAppErrorMsg, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, AppSender_SendCrossChainAppError_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AppSenderServer is the server API for AppSender service.
// All implementations must embed UnimplementedAppSenderServer
// for forward compatibility
type AppSenderServer interface {
	SendAppRequest(context.Context, *SendAppRequestMsg) (*emptypb.Empty, error)
	SendAppResponse(context.Context, *SendAppResponseMsg) (*emptypb.Empty, error)
	SendAppError(context.Context, *SendAppErrorMsg) (*emptypb.Empty, error)
	SendAppGossip(context.Context, *SendAppGossipMsg) (*emptypb.Empty, error)
	SendCrossChainAppRequest(context.Context, *SendCrossChainAppRequestMsg) (*emptypb.Empty, error)
	SendCrossChainAppResponse(context.Context, *SendCrossChainAppResponseMsg) (*emptypb.Empty, error)
	SendCrossChainAppError(context.Context, *SendCrossChainAppErrorMsg) (*emptypb.Empty, error)
	mustEmbedUnimplementedAppSenderServer()
}

// UnimplementedAppSenderServer must be embedded to have forward compatible implementations.
type UnimplementedAppSenderServer struct {
}

func (UnimplementedAppSenderServer) SendAppRequest(context.Context, *SendAppRequestMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendAppRequest not implemented")
}
func (UnimplementedAppSenderServer) SendAppResponse(context.Context, *SendAppResponseMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendAppResponse not implemented")
}
func (UnimplementedAppSenderServer) SendAppError(context.Context, *SendAppErrorMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendAppError not implemented")
}
func (UnimplementedAppSenderServer) SendAppGossip(context.Context, *SendAppGossipMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendAppGossip not implemented")
}
func (UnimplementedAppSenderServer) SendCrossChainAppRequest(context.Context, *SendCrossChainAppRequestMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCrossChainAppRequest not implemented")
}
func (UnimplementedAppSenderServer) SendCrossChainAppResponse(context.Context, *SendCrossChainAppResponseMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCrossChainAppResponse not implemented")
}
func (UnimplementedAppSenderServer) SendCrossChainAppError(context.Context, *SendCrossChainAppErrorMsg) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCrossChainAppError not implemented")
}
func (UnimplementedAppSenderServer) mustEmbedUnimplementedAppSenderServer() {}

// UnsafeAppSenderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AppSenderServer will
// result in compilation errors.
type UnsafeAppSenderServer interface {
	mustEmbedUnimplementedAppSenderServer()
}

func RegisterAppSenderServer(s grpc.ServiceRegistrar, srv AppSenderServer) {
	s.RegisterService(&AppSender_ServiceDesc, srv)
}

func _AppSender_SendAppRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendAppRequestMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendAppRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendAppRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendAppRequest(ctx, req.(*SendAppRequestMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppSender_SendAppResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendAppResponseMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendAppResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendAppResponse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendAppResponse(ctx, req.(*SendAppResponseMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppSender_SendAppError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendAppErrorMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendAppError(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendAppError_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendAppError(ctx, req.(*SendAppErrorMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppSender_SendAppGossip_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendAppGossipMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendAppGossip(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendAppGossip_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendAppGossip(ctx, req.(*SendAppGossipMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppSender_SendCrossChainAppRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendCrossChainAppRequestMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendCrossChainAppRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendCrossChainAppRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendCrossChainAppRequest(ctx, req.(*SendCrossChainAppRequestMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppSender_SendCrossChainAppResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendCrossChainAppResponseMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendCrossChainAppResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendCrossChainAppResponse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendCrossChainAppResponse(ctx, req.(*SendCrossChainAppResponseMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppSender_SendCrossChainAppError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendCrossChainAppErrorMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppSenderServer).SendCrossChainAppError(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AppSender_SendCrossChainAppError_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppSenderServer).SendCrossChainAppError(ctx, req.(*SendCrossChainAppErrorMsg))
	}
	return interceptor(ctx, in, info, handler)
}

// AppSender_ServiceDesc is the grpc.ServiceDesc for AppSender service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AppSender_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "appsender.AppSender",
	HandlerType: (*AppSenderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendAppRequest",
			Handler:    _AppSender_SendAppRequest_Handler,
		},
		{
			MethodName: "SendAppResponse",
			Handler:    _AppSender_SendAppResponse_Handler,
		},
		{
			MethodName: "SendAppError",
			Handler:    _AppSender_SendAppError_Handler,
		},
		{
			MethodName: "SendAppGossip",
			Handler:    _AppSender_SendAppGossip_Handler,
		},
		{
			MethodName: "SendCrossChainAppRequest",
			Handler:    _AppSender_SendCrossChainAppRequest_Handler,
		},
		{
			MethodName: "SendCrossChainAppResponse",
			Handler:    _AppSender_SendCrossChainAppResponse_Handler,
		},
		{
			MethodName: "SendCrossChainAppError",
			Handler:    _AppSender_SendCrossChainAppError_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "appsender/appsender.proto",
}
