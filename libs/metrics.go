package libs

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
)

func StartMetric(ctx context.Context, server *grpc.Server, metricPort int) {
	grpc_prometheus.Register(server)
	grpc_prometheus.EnableHandlingTimeHistogram()
	if metricPort != 0 {
		http.Handle("/metrics", promhttp.Handler())
		go http.ListenAndServe(":"+strconv.Itoa(metricPort), nil)
	}
}

func StartStatusMetric(ctx context.Context, metricPort int) {
	if metricPort != 0 {
		http.Handle("/metrics", promhttp.Handler())
		go http.ListenAndServe(":"+strconv.Itoa(metricPort), nil)
	}
}
