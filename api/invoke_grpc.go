/**
* @Author: TongTongLiu
* @Date: 2021/3/10 11:41 AM
**/

package api

import (
	"api-test/configs"
	"api-test/payload"
	"bytes"
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"github.com/fullstorydev/grpcurl"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type GRPCInterface struct {
	// interface
	Interface
	InterfaceNew

	// struct
	InterfaceRequest

	ReqPath     string
	ImportPaths []string
	ProtoFiles  []string
	MethodName  string
}

func (g *GRPCInterface) describeHeader(ctx context.Context, request *payload.Request) (headers []string) {
	if request.Header == nil || len(request.Header) == 0 {
		return
	}
	for k, v := range request.Header {
		headers = append(headers, k+":"+v)
	}
	return
}

func (g *GRPCInterface) Invoke(ctx context.Context, request *payload.Request, environment string) (pResp *payload.Response, err error) {
	if g.ProtoFiles == nil || len(g.ProtoFiles) == 0 {
		err = status.Errorf(codes.InvalidArgument, "Not found Proto files")
		return
	}
	cc, err := grpc.Dial(g.ReqPath, grpc.WithInsecure())
	defer cc.Close()
	if err != nil {
		xzap.Logger(ctx).Info("connect to envoy failed", zap.Any("err", err))
		err = status.Errorf(codes.Unknown, "connect to envoy failed , err: %v", err)
		return
	}
	var protoFileRootPath = configs.GetGlobalConfig().ProtoFile.RootPath
	var NewImportPaths = []string{
		protoFileRootPath,
		fmt.Sprintf("%s/protos/", protoFileRootPath),
		fmt.Sprintf("%s/%s", protoFileRootPath, g.ProtoFiles[0]),
	}
	var descSource grpcurl.DescriptorSource
	for i := 0; i < 3; i++ {
		descSource, err = grpcurl.DescriptorSourceFromProtoFiles(NewImportPaths, g.ProtoFiles...)
		if err == nil {
			break
		}
		xzap.Logger(ctx).Error("DescriptorSourceFromProtoFiles failed",
			zap.Any("err", err))
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		err = status.Errorf(codes.Unknown, "DescriptorSourceFromProtoFiles failed , err: %v", err)
		return
	}
	headers := g.describeHeader(ctx, request)
	rf, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.FormatJSON, descSource,
		true, true, strings.NewReader(string(request.Body)))
	if err != nil {
		xzap.Logger(ctx).Error("RequestParserAndFormatterFor failed",
			zap.Any("err", err))
		err = status.Errorf(codes.Unknown, "RequestParserAndFormatterFor failed, err: %v", err)
		return
	}
	var respBuf bytes.Buffer
	beginTime := time.Now()

	h := &grpcurl.DefaultEventHandler{
		Out:            &respBuf,
		Formatter:      formatter,
		VerbosityLevel: 1,
	}
	pResp = new(payload.Response)
	err = grpcurl.InvokeRPC(context.Background(), descSource, cc, g.MethodName, headers, h, rf.Next)
	endTime := time.Now()
	pResp.TimeCost = int(endTime.Sub(beginTime).Nanoseconds() / 1e6)
	if err != nil {
		xzap.Logger(ctx).Error("InvokeRPC failed",
			zap.Any("err", err))
		return
	}

	if h != nil && h.Status != nil {
		pResp.StatusCode = int(h.Status.Code())
	}
	if h.Status.Code() != codes.OK {
		pResp.Body = []byte(h.Status.Message())
		return
	}
	respString := respBuf.String()
	fmt.Println("respString", respString)
	strHeaderPos := strings.Index(respString, verboseResponseHeader)
	strContentPos := strings.Index(respString, verboseResponseContents)
	strTrailerPos := strings.Index(respString, verboseResponseTrailer)
	if strHeaderPos <= 0 || strContentPos <= 0 || strTrailerPos <= 0 {
		return
	}
	headerStr := respString[strHeaderPos+28 : strContentPos-1]
	headerList := strings.Split(headerStr, "\n")
	pRespHeader := make(map[string]string)
	for _, h := range headerList {
		hList := strings.Split(h, ": ")
		pRespHeader[hList[0]] = hList[1]
	}
	bodyStr := respString[strContentPos+20 : strTrailerPos]

	pResp.Header = pRespHeader
	pResp.Body = []byte(bodyStr)

	return
}
