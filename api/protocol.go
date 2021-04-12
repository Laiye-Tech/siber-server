package api

import (
	"api-test/dao"
	"api-test/siberconst"
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const (
	EnvironmentTest  string = "test"
	EnvironmentDev   string = "dev"
	EnvironmentStage string = "stage"
	EnvironmentProd  string = "prod"
)

func GetProtocolURL(ctx context.Context, protocolType string, environmentId string, environmentName string) (url string, err error) {
	EnvInfo := &siber.EnvInfo{
		EnvId: environmentId,
	}
	envInfoOutput, err := dao.NewDao().SelectEnv(ctx, EnvInfo)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "env is null")
		return
	}

	switch protocolType {
	// 根据传输协议选择地址
	case siberconst.GRPCProtocol:
		// gRPC
		switch environmentName {
		case EnvironmentDev:
			url = envInfoOutput.Grpc.DevEnvoy
		case EnvironmentTest:
			url = envInfoOutput.Grpc.TestEnvoy
		case EnvironmentStage:
			url = envInfoOutput.Grpc.StageEnvoy
		case EnvironmentProd:
			url = envInfoOutput.Grpc.ProdEnvoy
		}
		http := strings.HasPrefix(url, "http://")
		https := strings.HasPrefix(url, "https://")
		if http || https {
			err = status.Errorf(codes.InvalidArgument, "URL Incorrect format,please check")
			return
		}
	case siberconst.HTTPProtocol:
		// HTTP
		switch environmentName {
		case EnvironmentDev:
			url = envInfoOutput.Http.DevUrl
		case EnvironmentTest:
			url = envInfoOutput.Http.TestUrl
		case EnvironmentStage:
			url = envInfoOutput.Http.StageUrl
		case EnvironmentProd:
			url = envInfoOutput.Http.ProdUrl
		}
		http := strings.HasPrefix(url, "http://")
		https := strings.HasPrefix(url, "https://")
		if !(http || https) {
			err = status.Errorf(codes.InvalidArgument, "URL Incorrect format,please check")
			return
		}
	}
	if url == "" {
		err = status.Errorf(codes.InvalidArgument, "URL is nil,please check")
		return
	}
	return
}
