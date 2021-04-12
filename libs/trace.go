package libs

import (
	"api-test/configs"
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"runtime"
	"sync"
	"time"
)

type Tracer interface {
	GetOpenTracer() opentracing.Tracer
	StartFromCtx(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (Spaner, context.Context)
}

type Spaner interface {
	SetTag(string, interface{}) Spaner
	Finish()
}

var (
	tracerInstances sync.Map
)

type FakeTrace struct{}

func (f *FakeTrace) GetOpenTracer() opentracing.Tracer {
	return nil
}
func (f *FakeTrace) StartFromCtx(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (Spaner, context.Context) {
	return &FakeSpan{}, ctx
}
func NewFakeTrace(name string) *FakeTrace {
	return &FakeTrace{}
}

type FakeSpan struct{}

func (f *FakeSpan) SetTag(string, interface{}) Spaner { return f }
func (f *FakeSpan) Finish()                           {}

type Span struct {
	span opentracing.Span
}

func (s *Span) SetTag(k string, v interface{}) Spaner {
	s.span = s.span.SetTag(k, v)
	return s
}

func (s *Span) Finish() {
	s.span.Finish()
}

type Trace struct {
	tracer  opentracing.Tracer
	span    opentracing.Span
	success bool
	closer  io.Closer
}

func NewTrace(name string) *Trace {

	serviceName := configs.GetGlobalConfig().Jaeger.ServiceName
	var tracer *Trace
	agent := fmt.Sprintf("%v:%v", configs.GetInternalIp(), configs.GetGlobalConfig().Jaeger.AgentPort)
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			QueueSize:           1,
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agent,
		},
	}
	Log().Info(context.Background(), "new trace agent: %+v", agent)
	var logger jaegerlog.Logger
	logger = jaegerlog.StdLogger
	if !configs.AppDebug() {
		logger = jaegerlog.NullLogger
	}
	t, closer, err := cfg.New(
		serviceName,
		config.Logger(logger),
	)

	if err != nil {
		tracer = &Trace{
			success: false,
		}
		Log().Error(context.Background(), "init trace error (%v)", err)
	} else {
		opentracing.SetGlobalTracer(t)
		tracer = &Trace{
			tracer:  t,
			success: true,
			closer:  closer,
		}
	}
	return tracer
}

func GetTracer(name string) Tracer {
	var tracer Tracer
	if tracer, ok := tracerInstances.Load(name); ok {
		return tracer.(Tracer)
	}
	if !configs.GetGlobalConfig().Jaeger.Disable {
		tracer = NewTrace(name)
	} else {
		tracer = NewFakeTrace(name)
	}
	tracerInstances.Store(name, tracer)
	return tracer
}

func (t *Trace) GetOpenTracer() opentracing.Tracer {
	return t.tracer
}

func (t *Trace) StartFromCtx(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (Spaner, context.Context) {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}
	span := t.tracer.StartSpan(operationName, opts...)
	return &Span{span: span}, opentracing.ContextWithSpan(ctx, span)
}

func OpenTraceServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	tracer := GetTracer("default")
	if tracer.GetOpenTracer() != nil {
		if configs.GetGlobalConfig().Jaeger.Payload {
			return otgrpc.OpenTracingServerInterceptor(tracer.GetOpenTracer(), otgrpc.LogPayloads())(ctx, req, info, handler)
		} else {
			return otgrpc.OpenTracingServerInterceptor(tracer.GetOpenTracer())(ctx, req, info, handler)
		}
	}
	return handler(ctx, req)
}

func OpenTraceClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	tracer := GetTracer("default")
	if tracer.GetOpenTracer() != nil {
		if configs.GetGlobalConfig().Jaeger.Payload {
			return otgrpc.OpenTracingClientInterceptor(tracer.GetOpenTracer(), otgrpc.LogPayloads())(ctx, method, req, reply, cc, invoker, opts...)
		} else {
			return otgrpc.OpenTracingClientInterceptor(tracer.GetOpenTracer())(ctx, method, req, reply, cc, invoker, opts...)
		}
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func GetTraceOpName() string {
	f, _, _, callOk := runtime.Caller(1)
	if callOk {
		return runtime.FuncForPC(f).Name()
	}
	return ""
}
