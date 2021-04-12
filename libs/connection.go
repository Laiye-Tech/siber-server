package libs

import (
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

// Connection defines interface of connection
type Connection interface {
	GetGrpcConnection(ctx context.Context) *grpc.ClientConn
	Connect(ctx context.Context, address ...string)
}

// EnvoyResolverConnection helps connecting envoy
type EnvoyResolverConnection struct {
	grpcConn *grpc.ClientConn
	address  string
}

// NewEnvoyResolverConnection creates envoy connection
func NewEnvoyResolverConnection(address string) *EnvoyResolverConnection {
	return &EnvoyResolverConnection{
		address: address,
	}
}

// Connect connects envoy
func (e *EnvoyResolverConnection) Connect(ctx context.Context, address ...string) {
	ctx, cancel1 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel1()
	var conn *grpc.ClientConn
	var err error
	xzap.Logger(ctx).Info("envoy connect", zap.Any("address", address))
	conn, err = grpc.DialContext(ctx,
		e.address,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			PermitWithoutStream: true,
			Time:                5 * time.Minute,
		}),
		grpc.WithUnaryInterceptor(OpenTraceClientInterceptor),
	)
	if err != nil {
		xzap.Logger(ctx).Error("dial service(%s) by envoy resolver server error (%v)",
			zap.String("address", e.address),
			zap.String("err", err.Error()))
	}
	e.grpcConn = conn
}

// GetGrpcConnection gets active connection for request
func (e *EnvoyResolverConnection) GetGrpcConnection(ctx context.Context) *grpc.ClientConn {
	if e.grpcConn == nil || e.grpcConn.GetState() == connectivity.Shutdown {
		e.Connect(ctx)
	}
	return e.grpcConn
}
