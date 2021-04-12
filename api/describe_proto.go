/**
* @Author: TongTongLiu
* @Date: 2019/12/6 11:05 上午
**/

package api

import (
	"api-test/configs"
	"api-test/dao"
	"api-test/siberconst"
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 根据proto中定义的message，描述request的初始值

func describeMethodRequest(ctx context.Context, method *Method) (result string, err error) {
	if method == nil {
		err = status.Errorf(codes.InvalidArgument, "describe nil method")
		return
	}

	g := method.Interfaces[siberconst.GRPCProtocol]
	// 没有grpc方法，没有proto文件则不填充request
	if g == nil {
		return
	}
	var protoFile string
	rootPath := configs.GetGlobalConfig().ProtoFile.RootPath
	if grpcInterface, ok := g.(*GRPCInterface); !ok {
		err = status.Errorf(codes.InvalidArgument, "describeMethod got wrong interface type")
		return
	} else {
		if grpcInterface.ProtoFiles == nil || grpcInterface.ProtoFiles[0] == "" {
			err = status.Errorf(codes.InvalidArgument, "proto file is nil")
			return
		}
		protoFile = rootPath + "/" + grpcInterface.ProtoFiles[0]
	}
	ds, err := getFileDescribe(ctx, protoFile)
	if err != nil {
		return
	}
	dsc, err := ds.source.FindSymbol(method.Name)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "describeMethod FindSymbol failed, err %v", err)
		return
	}
	if methodDsc, ok := dsc.(*desc.MethodDescriptor); !ok {
		err = status.Errorf(codes.InvalidArgument, "describeMethod dsc wrong type")
		return
	} else {
		descInput := methodDsc.GetInputType()
		message := grpcurl.MakeTemplate(descInput)
		jsm := jsonpb.Marshaler{EmitDefaults: true}
		result, err = jsm.MarshalToString(message)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "jsm MarshalToString failed, err:%v", err)
			return
		}
	}
	return
}

func DescribeRequest(ctx context.Context, request *siber.MethodInfo) (response *siber.MethodDescribe, err error) {
	if request == nil || request.MethodName == "" {
		err = status.Errorf(codes.InvalidArgument, "DescribeRequest failed, get nil request method name")
		return
	}
	methodInfo, err := dao.NewDao().SelectMethod(ctx, request)
	if err != nil {
		return
	}
	var result string
	switch methodInfo.MethodType {
	case siberconst.GRPCMethod:
		if methodInfo == nil || len(methodInfo.ProtoFiles) == 0 {
			err = status.Errorf(codes.InvalidArgument, "DescribeRequest failed, get nil proto file ")
			return
		}
		interfaceGRPC := &GRPCInterface{
			Interface:   nil,
			ImportPaths: methodInfo.ImportPaths,
			ProtoFiles:  methodInfo.ProtoFiles,
		}
		method := &Method{
			Name: methodInfo.MethodName,
			Interfaces: map[string]Interface{
				siberconst.GRPCProtocol: interfaceGRPC,
			},
		}
		result, err = describeMethodRequest(ctx, method)

	case siberconst.GraphQLMethod:
		// TODO:Q2 子杰会已接口形式提供标准格式
		result = "{}"
	case siberconst.HTTPMethod:
		// TODO:Q2 接入模板渲染
		result = "{}"
	default:
		err = status.Errorf(codes.InvalidArgument, "unsupported method type :%s", methodInfo.MethodType)
	}
	if err != nil {
		return
	}
	response = &siber.MethodDescribe{
		MethodName:     request.MethodName,
		RequestMessage: result,
	}
	return
}
