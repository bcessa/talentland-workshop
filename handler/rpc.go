package handler

import (
	"context"

	protov1 "github.com/bcessa/echo-service/proto/sample/v1"
	"go.bryk.io/pkg/net/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type rpcInterface struct {
	so *ServiceOperator
	protov1.UnimplementedServiceAPIServer
}

// RPC can be used to expose the service functionality through
// a gRPC/HTTP server instance.
func (so *ServiceOperator) RPC() rpc.ServiceProvider {
	return &rpcInterface{so: so}
}

func (r *rpcInterface) Ping(_ context.Context, _ *emptypb.Empty) (*protov1.PingResponse, error) {
	return &protov1.PingResponse{Ok: r.so.Ping()}, nil
}

func (r *rpcInterface) Ready(_ context.Context, _ *emptypb.Empty) (*protov1.ReadyResponse, error) {
	return &protov1.ReadyResponse{Ok: r.so.Ready()}, nil
}

func (r *rpcInterface) Echo(ctx context.Context, req *protov1.EchoRequest) (*protov1.EchoResponse, error) {
	res, err := r.so.Echo(ctx, req.Value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &protov1.EchoResponse{
		Result: res,
	}, nil
}

func (r *rpcInterface) Faulty(ctx context.Context, _ *emptypb.Empty) (*protov1.DummyResponse, error) {
	if err := r.so.Faulty(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &protov1.DummyResponse{Ok: true}, nil
}

func (r *rpcInterface) Slow(ctx context.Context, _ *emptypb.Empty) (*protov1.DummyResponse, error) {
	if err := r.so.Slow(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &protov1.DummyResponse{Ok: true}, nil
}

func (r *rpcInterface) ServerSetup(server *grpc.Server) {
	protov1.RegisterServiceAPIServer(server, r)
}

func (r *rpcInterface) ServiceDesc() grpc.ServiceDesc {
	return protov1.ServiceAPI_ServiceDesc
}

func (r *rpcInterface) GatewaySetup() rpc.GatewayRegisterFunc {
	return protov1.RegisterServiceAPIHandler
}
