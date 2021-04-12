/**
* @Author: TongTongLiu
* @Date: 2019-08-07 20:58
**/

package libs

import (
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	TraceIdKey = "trace-id"
)

func ExtractTraceId(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range md {
			if k == TraceIdKey && len(v) > 0 {
				return v[0]
			}
		}
	}
	return ""
}

//内部的context全用incoming，调用外部的grpc服务时，统一转成outcoming
func AddInnerTraceId(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	extractId := ExtractTraceId(ctx)
	//traceId 为空 增加traceId
	if (!ok) || (ok && extractId == "") {
		traceId, _ := uuid.NewV4()
		trace := metadata.Pairs("trace-id", traceId.String())
		md = metadata.Join(md, trace)
		ctx = metadata.NewIncomingContext(ctx, md)
	}
	return ctx
}

func ChangeInComingCxtToOut(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {

		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// ExtractTraceID extracts trace id from context
func ExtractTraceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range md {
			if k == TraceIdKey && len(v) > 0 {
				return v[0]
			}
		}
	}
	return ""
}

//AddInnerTraceID 内部的context全用incoming，调用外部的grpc服务时，统一转成outcoming
func AddInnerTraceID(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	extractID := ExtractTraceID(ctx)
	//traceID 为空 增加traceID
	if (!ok) || (ok && extractID == "") {
		traceID, _ := uuid.NewV4()
		trace := metadata.Pairs("trace-id", traceID.String())
		md = metadata.Join(md, trace)
		ctx = metadata.NewIncomingContext(ctx, md)
	}
	return ctx
}

// AddMethodNameInterceptor adds method name in context
func AddTraceIdInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error
	ctx = AddInnerTraceID(ctx)
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, err
}
