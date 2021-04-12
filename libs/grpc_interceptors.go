package libs

import (
	"api-test/configs"
	"fmt"
	"github.com/deckarep/golang-set"
	//"github.com/golang/protobuf/jsonpb"
	//"github.com/golang/protobuf/proto"
	"github.com/palantir/stacktrace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
	"runtime"
	"strings"
	"time"
)

var (
	traceKeys        mapset.Set
	contextTraceKeys = []string{"user-agent", "trace-id"}
	probability      = 100
	//marshal          = jsonpb.Marshaler{}
)

func init() {
	traceKeys = mapset.NewSet()
	traceKeys.Add("trace-id")
	for _, key := range contextTraceKeys {

		traceKeys.Add(key)

	}
}

func getClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}

func GetClientIP(ctx context.Context) (string, error) {
	return getClientIP(ctx)
}

func RequestTraceInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTs := time.Now().UnixNano()
	//reqStr, _ := marshal.MarshalToString(req.(proto.Message))
	resp, err := handler(ctx, req)
	endTs := time.Now().UnixNano()
	r := NewRandom().GetRandomInt(100)
	needLog := true
	if r > probability {
		needLog = false
	}
	if needLog {
		p, ok := peer.FromContext(ctx)
		md, ok1 := metadata.FromIncomingContext(ctx)
		traceInfo := make([]string, 0, traceKeys.Cardinality()+1)
		if ok1 {
			for k, v := range md {
				if traceKeys.Contains(k) {
					traceInfo = append(traceInfo, fmt.Sprintf("%v(%v)", k, strings.Join(v, ",")))
				}

			}
		}
		if ok {
			funcName, file, line, callOk := runtime.Caller(5)
			requestInfo := &grpc.UnaryServerInfo{}
			fileData := strings.Split(file, "/")
			if configs.AppDebug() {
				requestInfo = info
			}
			if callOk {
				var trace error
				if err != nil {
					trace = stacktrace.Propagate(err, "")
				}
				Log().Info(ctx, "%v [line:%v] %v() address(%v) traceInfo:{%v} requestInfo:{%+v} err:{%+v} trace:{%+v} cost (%v) ns",
					fileData[len(fileData)-1], line, runtime.FuncForPC(funcName).Name(), p.Addr, strings.Join(traceInfo, " "), requestInfo, err, trace, endTs-startTs)
			}
		}
	}

	return resp, err
}

func AddMethodNameInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	Log().Debug(ctx, "add method name(%v)", info.FullMethod)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{"method": info.FullMethod})
	} else {
		md["method"] = []string{info.FullMethod}
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return handler(ctx, req)
}
