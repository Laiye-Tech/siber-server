/**
* @Author: TongTongLiu
* @Date: 2019-09-12 12:27
**/

package main

import (
	"api-test/configs"
	"api-test/core"
	"api-test/initial"
	"api-test/libs"
	"api-test/service"
	"api-test/sibercron"
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func runGateWay(endPoint string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true}),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}
	registers := []func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error{
		siber.RegisterSiberServiceHandlerFromEndpoint,
	}
	for _, r := range registers {
		err := r(ctx, mux, endPoint, opts)
		if err != nil {
			xzap.Logger(ctx).Error("start siber proxy gate way error", zap.Any("err", err))
			fmt.Println(fmt.Sprintf("start siber proxy gate way error (%v)", err))
			panic(err)
		}
	}

	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", configs.FlagGWPort),
		Handler: mux,
	}
	err2 := s.ListenAndServe()
	xzap.Logger(ctx).Info("grpc gate way start ", zap.Any("port", configs.FlagGWPort), zap.Any("err", err2))
	return err2
}

func loop() {
	c := cron.New(cron.WithSeconds())

	_, err := c.AddFunc("23 23 01 * * ?", func() {
		sibercron.SetServiceRoutes(context.Background())
	})
	if err != nil {
		xzap.Logger(context.Background()).Error("Add func SetServiceRoutes failed ", zap.Any("error", err))
	}

	_, err = c.AddFunc("23 23 02 * * ?", func() {
		_ = sibercron.RefreshAllMethodServices(context.Background())
	})
	if err != nil {
		xzap.Logger(context.Background()).Error("Add func RefreshAllMethodServices failed ", zap.Any("error", err))
	}

	_, err = c.AddFunc("23 23 03 * * ?", func() {
		_ = sibercron.RefreshAllPlanServices(context.Background())
	})
	if err != nil {
		xzap.Logger(context.Background()).Error("Add func RefreshAllPlanServices failed ", zap.Any("error", err))
	}
}
func cronStat() {
	err := sibercron.CaseLogNumCron(context.Background())
	if err != nil {
		xzap.Logger(context.Background()).Error("Init CaseLogNumCron failed ", zap.Any("error", err))
	}
	err = sibercron.PlanLogNumCron(context.Background())
	if err != nil {
		xzap.Logger(context.Background()).Error("Init PlanLogNumCron failed ", zap.Any("error", err))
	}

	c := cron.New(cron.WithSeconds())
	_, err = c.AddFunc("0 0 2 * * ?", func() {
		_ = sibercron.CaseLogNumCron(context.Background())
	})
	if err != nil {
		xzap.Logger(context.Background()).Error("Add func CaseLogNumCron failed ", zap.Any("error", err))
	}
	_, err = c.AddFunc("0 0 3 * * ?", func() {
		_ = sibercron.PlanLogNumCron(context.Background())
	})
	if err != nil {
		xzap.Logger(context.Background()).Error("Add func PlanLogNumCron failed ", zap.Any("error", err))
	}
}
func main() {
	initial.Initial()
	configs.FlagParse()
	port := configs.FlagPort
	if port == 0 {
		fmt.Println("port must be given")
		os.Exit(1)
	}
	endPoint := fmt.Sprintf(":%v", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	//加入token验证
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				// libs.UnaryServerUserAuthCheckIntercepter,
				libs.RequestTraceInterceptor,
				libs.OpenTraceServerInterceptor,
				libs.AddTraceIdInterceptor,
			),
		),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				MaxConnectionIdle: time.Hour,
			},
		),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
	)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	siber.RegisterSiberServiceServer(s, service.NewSiberService())
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	reflection.Register(s)
	libs.StartMetric(ctx, s, configs.GetGlobalConfig().Port.Prometheus)

	sibercron.SetServiceRoutes(context.Background())
	loop()

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		} else {
			fmt.Printf("grpc service start at port (%v)", port)
			xzap.Logger(ctx).Info("grpc server at port", zap.Any("port", port))
		}
	}()
	go func() {
		if err1 := runGateWay(endPoint); err1 != nil {
			os.Exit(1)
		}

	}()
	go func() {
		err = core.ScheduleManagerGlobal.Run(ctx)
	}()
	go func() {
		cronStat()
	}()
	sign := <-signalChan

	xzap.Logger(ctx).Info("grpc server will stop", zap.Any("signal", sign))
	s.GracefulStop()
	cancel()
}
