package api

import (
	"api-test/configs"
	"api-test/dao"
	"api-test/payload"
	"api-test/sibercron"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const WebLineBreak = "\n###"

const (
	CreateItemMode    = "CREATE"
	UpdateItemMode    = "UPDATE"
	DeleteItemMode    = "DELETE"
	QueryItemMode     = "QUERY"
	DuplicateItemMode = "DUPLICATE"
	ListMode          = "List"
)

type Method struct {
	Name         string
	ProtoFile    *descriptor.FileDescriptorProto
	ProtoService *descriptor.ServiceDescriptorProto
	ProtoMethod  *descriptor.MethodDescriptorProto

	Interfaces map[string]Interface
}

type descSourceCase struct {
	name        string
	source      grpcurl.DescriptorSource
	includeRefl bool
}

func (m *Method) Invoke(ctx context.Context, protocol string, request *payload.Request, environment string) (pResp *payload.Response, err error) {
	if request == nil {
		err = status.Errorf(codes.InvalidArgument, "request is nil")
		return
	}
	pResp, err = m.Interfaces[protocol].Invoke(ctx, request, environment)
	return
}

func getImportPath(ctx context.Context, fileName string) (importPaths []string, err error) {
	protoFileRootPath := configs.GetGlobalConfig().ProtoFile.RootPath
	s := strings.Split(fileName, "/")
	if len(s) == 0 {
		err = status.Errorf(codes.InvalidArgument, "nil is invalid proto file")
		return
	}
	filePath := ""
	if len(s) == 1 {
		filePath = ""
	} else {
		filePath = strings.Join(s[:len(s)-1], "/")
	}

	importPaths = []string{
		protoFileRootPath,
		fmt.Sprintf("%s/protos/", protoFileRootPath),
		fmt.Sprintf("%s/%s", protoFileRootPath, filePath),
	}
	return
}

func getFileDescribe(ctx context.Context, fileName string) (ds descSourceCase, err error) {
	importPaths, err := getImportPath(ctx, fileName)
	if err != nil {
		return
	}
	otherSourceProtoFiles, err := grpcurl.DescriptorSourceFromProtoFiles(importPaths, fileName)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "describe proto %s failed :%v", fileName, err)
		return

	}
	ds = descSourceCase{"proto", otherSourceProtoFiles, false}

	return
}

// 通过ProtoFile中获取包含的所有service
func ParseServiceFromProtoFile(ctx context.Context, protoFile string) (serviceList []string, err error) {
	ds, err := getFileDescribe(ctx, protoFile)
	if err != nil {
		return
	}
	serviceList, err = grpcurl.ListServices(ds.source)
	if err != nil {
		fmt.Printf("list services failed :%v", err)
		return
	}
	return
}

// 通过ProtoFile获取service中包含的method
func ParseMethodsFromProtoFile(ctx context.Context, protoFile string, serviceList []string) (methodList *siber.ServiceInfo, err error) {
	ds, err := getFileDescribe(ctx, protoFile)
	if err != nil {
		return
	}
	serviceList, err = ParseServiceFromProtoFile(ctx, protoFile)
	if err != nil {
		err = status.Errorf(codes.Unknown, "parse service from proto %s failed : %+v", protoFile, err)
		return
	}
	if len(serviceList) == 0 {
		err = status.Errorf(codes.InvalidArgument, "Can't find services")
		return
	}
	var methods []string
	for _, s := range serviceList {
		if s == "" {
			err = status.Errorf(codes.InvalidArgument, "service name is null")
			return
		}
		m, err := grpcurl.ListMethods(ds.source, s)
		if err != nil {
			err = status.Errorf(codes.Unknown, "list method for service %s from proto %s failed : %+v", s, protoFile, err)
			return nil, err
		}
		for _, mm := range m {
			methods = append(methods, mm)
		}
	}
	if len(methods) == 0 {
		return
	}
	methodList = new(siber.ServiceInfo)
	methodList.MethodList = methods
	return
}

// 获得方法描述，包括但不限于：名称，输入，输出
func DescribeMethodFromProtoFile(ctx context.Context, info *siber.MethodInfo) (methodSC *siber.MethodDescribe, err error) {
	if info == nil {
		err = status.Errorf(codes.InvalidArgument, "describe method faile, err: info is nil")
		return
	}
	ds, err := getFileDescribe(ctx, info.ProtoFiles[0])
	if err != nil {
		return
	}
	dsc, err := ds.source.FindSymbol(info.MethodName)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "list methods failed :%v", err)
		return
	}
	methodSCP, ok := dsc.(*desc.MethodDescriptor)
	if !ok {
		err = status.Errorf(codes.Unknown, "format methodSC failed: %+v", methodSC)
		return
	}
	methodSC = new(siber.MethodDescribe)
	importPaths := []string{info.ProtoFiles[0]}
	httpInfo := methodSCP.GetOptions().String()
	if httpInfo == "" {
		methodSC.HttpUri = ""
		err = status.Errorf(codes.OK, "the service Not supported http")
	} else {
		infoList := strings.Split(httpInfo, "{")
		if len(infoList) < 2 {
			err = status.Errorf(codes.OutOfRange, "describe method failed, err: invalided options, httpInfo:%s", httpInfo)
			return
		}
		infoList = strings.Split(infoList[1], `:"`)
		if len(infoList) < 2 {
			err = status.Errorf(codes.OutOfRange, "describe method failed, err: invalided options, httpInfo:%s", httpInfo)
			return
		}
		methodSC.HttpRequestMode = infoList[0]
		infoList = strings.Split(infoList[1], `"`)
		if len(infoList) < 2 {
			err = status.Errorf(codes.OutOfRange, "describe method failed, err: invalided options, httpInfo:%s", httpInfo)
			return
		}
		methodSC.HttpUri = infoList[0]
	}

	inputMsg := "message: " + methodSCP.GetInputType().GetName()
	inputMsg += beautyMsgFields(methodSCP.GetInputType())
	methodSC.RequestMessage = inputMsg

	outputMsg := "message: " + methodSCP.GetOutputType().GetName()
	outputMsg += beautyMsgFields(methodSCP.GetOutputType())
	methodSC.ResponseMessage = outputMsg

	methodSC.ImportPaths = importPaths
	methodSC.MethodName = methodSCP.GetName()
	return
}

func beautyMsgFields(m *desc.MessageDescriptor) (msgStr string) {
	if m == nil {
		return
	}
	MsgFields := m.GetFields()
	if MsgFields == nil || len(MsgFields) == 0 {
		return
	}
	msgStr = ""
	for _, f := range MsgFields {
		msgStr += WebLineBreak
		switch f.GetLabel().String() {
		case "LABEL_OPTIONAL":
		case "LABEL_REPEATED":
			msgStr = msgStr + "repeated "
		default:
		}
		msgStr = msgStr + f.GetType().String() + " "
		msgStr = msgStr + f.GetName() + " = "
		msgStr = msgStr + strconv.Itoa(int(f.GetNumber())) + ";"
	}
	return
}

func GetMethodList(ctx context.Context, request *siber.FilterInfo) (response *siber.MethodList, err error) {
	methodList, totalNum, err := dao.NewDao().ListMethod(ctx, request)
	if err != nil {
		return
	}
	response = &siber.MethodList{
		MethodList: *methodList,
		TotalNum:   uint32(totalNum),
	}
	return
}

func ManageMethodInfo(ctx context.Context, methodInput *siber.ManageMethodInfo) (methodOutput *siber.MethodInfo, err error) {
	if methodInput == nil || methodInput.MethodInfo == nil {
		return
	}
	var service string
	switch methodInput.ManageMode {
	case CreateItemMode:
		service, err = sibercron.GetServicesForMethod(ctx, methodInput.MethodInfo)
		if err != nil {
			return
		}
		methodInput.MethodInfo.Service = service
		methodOutput, err = dao.NewDao().InsertMethod(ctx, methodInput.MethodInfo)
	case QueryItemMode:
		methodOutput, err = dao.NewDao().SelectMethod(ctx, methodInput.MethodInfo)
	case UpdateItemMode:
		service, err = sibercron.GetServicesForMethod(ctx, methodInput.MethodInfo)
		if err != nil {
			return
		}
		methodInput.MethodInfo.Service = service
		methodOutput, err = dao.NewDao().UpdateMethod(ctx, methodInput.MethodInfo)
	}

	return methodOutput, err
}

func GetProtoFileList(ctx context.Context, requests *siber.ProtoFileListRequests) (response *siber.ProtoFileListResponse) {
	var files []string
	protoFileRootPath := configs.GetGlobalConfig().ProtoFile.RootPath
	protoFilePath := path.Join(protoFileRootPath, "protos")
	err := filepath.Walk(protoFilePath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".proto") {
			partPath, _ := filepath.Rel(protoFileRootPath, path)
			files = append(files, partPath)
		}
		return nil
	})
	if len(files) == 0 {
		err = status.Errorf(codes.Unknown, "find proto files failed, please check the path")
	}
	if err != nil {
		return
	}
	response = &siber.ProtoFileListResponse{
		ProtoFiles: files,
	}
	return response
}

func GetGraphqlMethods(ctx context.Context, requests *siber.GraphqlMethodListRequest) (response *siber.GraphqlMethodListResponse, err error) {
	response, err = dao.NewDao().ListGraphqlMethod(ctx, requests)
	return
}
func GetGraphqlQuery(ctx context.Context, requests *siber.MethodInfo) (response *siber.MethodInfo, err error) {
	response, err = dao.NewDao().GetGraphqlQuery(ctx, requests)
	return
}
