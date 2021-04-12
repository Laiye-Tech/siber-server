/**
* @Author: TongTongLiu
* @Date: 2019/11/28 12:28 下午
**/

// 这个文件用于描绘case
// 属于case的前置操作，包括但不限于：
//   - 填充 case 的 variable 和 function
//   - 将引用了 SiberAuth 等自带算法的 header 渲染为目标格式

package core

import (
	"api-test/api"
	"api-test/dao"
	"api-test/describe"
	"api-test/payload"
	"api-test/siberconst"
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/golang/protobuf/ptypes/struct"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"strings"
	"time"
)

const versionDelimiter = "V"
const (
	MysqlInstance = "mysql"
	MongoInstance = "mongo"
	RedisInstance = "redis"
)
const (
	GraphQuery         = "query"
	GraphOperationName = "operationName"
	GraphVariables     = "variables"
)

// 将弱类型的 content从 struct 转换为 interface
// server 中使用 interface作为标准类型流通
func StructToStandard(ctx context.Context, rawValue *structpb.Value) (value interface{}, err error) {
	if rawValue == nil {
		err = status.Errorf(codes.InvalidArgument, "convertStructToProto failed , rawValue is nil")
		return
	}
	switch rawValue.Kind.(type) {
	case *structpb.Value_NumberValue:
		value = rawValue.GetNumberValue()
	case *structpb.Value_BoolValue:
		value = rawValue.GetBoolValue()
	case *structpb.Value_StringValue:
		value = rawValue.GetStringValue()
	case *structpb.Value_ListValue:
		valueList := rawValue.GetListValue()
		if len(valueList.Values) == 0 {
			break
		}
		var value []interface{}
		for _, v := range valueList.Values {
			vs, err := StructToStandard(ctx, v)
			if err != nil {
				return nil, err
			}
			value = append(value, vs)
		}
		return value, nil
	default:
		err = status.Errorf(codes.InvalidArgument, "unsupported raw value type, %v", reflect.TypeOf(rawValue))
		return
	}
	return
}

// 将 interface 的类型转换为 struct
// 用于向前端返回
func StandardToStruct(ctx context.Context, value interface{}) (structValue *structpb.Value, err error) {
	if value == nil {
		err = status.Errorf(codes.InvalidArgument, "StandardToStruct failed,value is nil")
		return
	}
	structValue = new(structpb.Value)

	vString, ok := value.(string)
	if ok {
		structValue.Kind = &structpb.Value_StringValue{StringValue: vString}
		return
	}

	vFloat, ok := value.(float64)
	if ok {
		structValue.Kind = &structpb.Value_NumberValue{NumberValue: vFloat}
		return
	}

	vBool, ok := value.(bool)
	if ok {
		structValue.Kind = &structpb.Value_BoolValue{BoolValue: vBool}
		return
	}

	vList, ok := value.([]interface{})
	if ok {
		var valueList []*structpb.Value
		for _, v := range vList {
			val, err := StandardToStruct(ctx, v)
			if err != nil {
				return nil, err
			}
			valueList = append(valueList, val)
		}

		listValue := &structpb.ListValue{
			Values: valueList,
		}
		structValue.Kind = &structpb.Value_ListValue{ListValue: listValue}
	}
	return
}

// 处理case version 中的所有struct 为interface 类型
func CaseVersionToStandard(ctx context.Context, caseInput *siber.CaseVersionInfo) (caseStandard *dao.CaseVersionStandard, err error) {
	// 不需要报错，有些case就是没有checkpoint
	if caseInput == nil {
		return
	}
	var checkPointList []*dao.CheckerStandard
	if caseInput.CheckPoint != nil {
		for _, c := range caseInput.CheckPoint {
			content, err := StructToStandard(ctx, c.Content)
			if err != nil {
				return nil, err
			}
			checkPoint := &dao.CheckerStandard{
				Key:      c.Key,
				Relation: c.Relation,
				Content:  content,
			}
			checkPointList = append(checkPointList, checkPoint)
		}
	}
	caseStandard = &dao.CaseVersionStandard{
		CaseId:         caseInput.CaseId,
		VersionControl: caseInput.VersionControl,
		RequestHeader:  caseInput.RequestHeader,
		UrlParameter:   caseInput.UrlParameter,
		RequestBody:    caseInput.RequestBody,
		CheckPoint:     checkPointList,
		InjectPoint:    caseInput.InjectPoint,
		SleepPoint:     caseInput.SleepPoint,
		Remark:         caseInput.Remark,
		InvalidDate:    caseInput.InvalidDate,
		UserUpdate:     caseInput.UserUpdate,
		InsertTime:     caseInput.InsertTime,
		UpdateTime:     caseInput.UpdateTime,
	}
	return
}

// 处理case version 中所有 interface 为struct 类型，用于向前端输出
func CaseVersionToProto(ctx context.Context, caseStandard *dao.CaseVersionStandard) (caseProto *siber.CaseVersionInfo, err error) {
	if caseStandard == nil {
		return
	}
	var checkPointList []*siber.CheckSub
	if caseStandard.CheckPoint != nil {
		for _, c := range caseStandard.CheckPoint {
			content, err := StandardToStruct(ctx, c.Content)
			if err != nil {
				return nil, err
			}
			checkPoint := &siber.CheckSub{
				Key:      c.Key,
				Relation: c.Relation,
				Content:  content,
			}
			checkPointList = append(checkPointList, checkPoint)
		}
	}
	caseProto = &siber.CaseVersionInfo{
		CaseId:         caseStandard.CaseId,
		VersionControl: caseStandard.VersionControl,
		RequestHeader:  caseStandard.RequestHeader,
		RequestBody:    caseStandard.RequestBody,
		UrlParameter:   caseStandard.UrlParameter,
		CheckPoint:     checkPointList,
		InjectPoint:    caseStandard.InjectPoint,
		SleepPoint:     caseStandard.SleepPoint,
		Remark:         caseStandard.Remark,
		InvalidDate:    caseStandard.InvalidDate,
		InsertTime:     caseStandard.InsertTime,
		UpdateTime:     caseStandard.UpdateTime,
		UserUpdate:     caseStandard.UserUpdate,
	}
	return
}

// 渲染 case 详情，包括：
//   - 自定义的鉴权算法，比如：SiberAuth
//   - function
//   - variable
func (c *Case) Render(ctx context.Context, variable *payload.Variable) (err error) {
	if c.Request.Header != nil && len(c.Request.Header) > 0 {
		var headers map[string]string
		headers, err = describe.DescCustomizeHeader(ctx, c.Request.Header, c.Plan.Trigger.Environment)
		if err != nil {
			return
		}
		c.Request.Header = headers
	}

	err = c.Request.Render(ctx, c.Flow.Variable)
	if err != nil {
		return err
	}
	return
}

func (c *Case) actionRender(ctx context.Context, variable *payload.Variable) (err error) {
	for i, _ := range c.Actions {
		err = c.Actions[i].Render(ctx, c.Flow.Variable)
		if err != nil {
			return
		}
	}
	return
}

// TODO: 不知道什么时候的代码了，没用的话可以删除
//func covertVersionTOFloat(ctx context.Context, versionString string) (versionFloat64 float64, err error) {
//	versionString = dao.VersionFormat(versionString)
//	l := strings.Split(versionString, versionDelimiter)
//	if len(l) < 2 {
//		err = status.Errorf(codes.OutOfRange, "unStandard version number: %s", versionString)
//		return
//	}
//	versionFloat64, err = strconv.ParseFloat(l[1], 64)
//	if err != nil {
//		err = status.Errorf(codes.OutOfRange, "unStandard version number: %s, err:%v", versionString, err)
//		return
//	}
//	return
//}

// 寻找不高于versionControl（plan级）的最高版本的case version info
func describeCaseVersion(ctx context.Context, c *Case) (caseVersionInfo *dao.CaseVersionStandard, err error) {
	// 对已有版本进行排序

	versionInfos, err := sortCaseVersion(ctx, c)
	if err != nil {
		return
	}
	length := len(versionInfos)
	if length == 0 {
		err = status.Errorf(codes.InvalidArgument, "case has no version, case id :%s", c.Id)
		return
	}
	caseVersionInput := &dao.CaseVersionStandard{
		CaseId: c.Id,
	}
	if c.Plan.Trigger.VersionControl == "" {
		// 如果plan版本为空，寻找最高版本的case
		caseVersionInput.VersionControl = versionInfos[0].CurrentVersion
	} else {
		caseVersionInput = &dao.CaseVersionStandard{
			CaseId:         c.Id,
			VersionControl: c.Plan.Trigger.VersionControl,
		}
		caseVersionInfo, err = dao.NewDao().SelectCaseVersion(ctx, caseVersionInput)
		if caseVersionInfo != nil && caseVersionInfo.CaseId != "" {
			return
		} else {
			caseVersionInput.VersionControl = ""
		}
		// 如果plan版本不为空，找不高于此版本的最高版本case
		planVersionNum, err := api.RegexVersion(c.Plan.Trigger.VersionControl)
		if err != nil {
			return nil, err
		}
		for _, v := range versionInfos {
			if v.AfterVersion > planVersionNum {
				continue
			}
			// 当前版本都是注入：V3.20 V5.01 这种，保留两个精度即可
			caseVersionInput.VersionControl = v.CurrentVersion
			break
		}
	}
	if caseVersionInput.VersionControl == "" {
		err = status.Errorf(codes.FailedPrecondition, "not valid version,plan version:%s, max case version:%s, min case version:%s",
			c.Plan.Trigger.VersionControl, versionInfos[0].CurrentVersion, versionInfos[length-1].CurrentVersion)
		return
	}
	caseVersionInfo, err = dao.NewDao().SelectCaseVersion(ctx, caseVersionInput)
	return
}

func describeCase(ctx context.Context, c *Case) (err error) {
	caseInfoInput := &siber.CaseInfo{
		CaseId: c.Id,
	}
	caseInfoOutput, err := dao.NewDao().SelectCase(ctx, caseInfoInput)
	if err != nil {
		return
	}
	caseVersionOutput, err := describeCaseVersion(ctx, c)
	if err != nil || caseVersionOutput == nil {
		return
	}
	c.Name = caseInfoOutput.CaseName
	c.CaseMode = caseInfoOutput.CaseMode
	c.Request = new(payload.Request)
	c.Request.Body = []byte(caseVersionOutput.RequestBody)
	c.Request.UrlParameter = caseVersionOutput.UrlParameter
	c.Request.Header = caseVersionOutput.RequestHeader
	switch c.CaseMode {
	case InjectCase:
		data := gjson.Get(caseVersionOutput.RequestBody, c.Plan.Trigger.Environment)
		if !data.Exists() {
			return
		}
		c.Request.Body = []byte(data.String())
	case InterfaceCase:
		methodInput := &siber.MethodInfo{
			MethodName: caseInfoOutput.MethodName,
		}
		methodOutput, err := dao.NewDao().SelectMethod(ctx, methodInput)
		if err != nil {
			return err
		}
		interfaceHTTP := &api.HTTPInterface{
			Interface: nil,
			ReqMode:   strings.ToUpper(methodOutput.HttpRequestMode),
			ReqPath:   fmt.Sprintf("%s%s", c.Plan.Url, methodOutput.HttpUri),
			ReqHeader: nil,
		}
		interfaceGRPC := &api.GRPCInterface{
			Interface:   nil,
			ImportPaths: methodOutput.ImportPaths,
			ProtoFiles:  methodOutput.ProtoFiles,
			ReqPath:     c.Plan.Url,
			MethodName:  methodOutput.MethodName,
		}
		c.Method = new(api.Method)
		c.Method.Name = methodOutput.MethodName
		c.Method.Interfaces = make(map[string]api.Interface)
		c.Method.Interfaces[siberconst.HTTPProtocol] = interfaceHTTP
		c.Method.Interfaces[siberconst.GRPCProtocol] = interfaceGRPC
	case InstanceCase:
		instance := &siber.EnvInfo{
			EnvName: caseInfoOutput.InstanceName,
		}
		instanceOutputInfo, err := dao.NewDao().SelectEnv(ctx, instance)
		if err != nil {
			return err
		}
		if instanceOutputInfo == nil || instanceOutputInfo.Instance == nil {
			err = status.Errorf(codes.FailedPrecondition, "instance info %s is nil", instanceOutputInfo.EnvName)
			return err
		}
		var client api.Instance
		switch instanceOutputInfo.Instance.InstanceType {
		case MysqlInstance:
			client = &api.MysqlClient{
				InstanceName: instance.EnvName,
			}
		default:
			err = status.Errorf(codes.InvalidArgument, "unsupported instance type: %s", instanceOutputInfo.Instance.InstanceType)
			return err
		}
		err = client.Init(ctx, instanceOutputInfo.Instance)
		if err != nil {
			return err
		}
		c.Instance = client
	}

	if c.Plan.Trigger == nil {
		err = status.Errorf(codes.InvalidArgument, "trigger is nil")
		return
	}
	c.Version = caseVersionOutput.VersionControl
	c.Actions = []CaseAction{}
	var errInfo error
	for _, injectPoint := range caseVersionOutput.InjectPoint {
		injectAction := new(InjectPoint)
		injectAction.VariableName = injectPoint.Content
		injectAction.Selector = createSelector(injectPoint.Key)
		c.Actions = append(c.Actions, injectAction)
	}
	for _, checkPoint := range caseVersionOutput.CheckPoint {
		checkAction := new(CheckPoint)
		checkAction.Selector = createSelector(checkPoint.Key)
		checkAction.Checker, err = createChecker(ctx, checkPoint)
		if err != nil {
			errInfo = err
		}
		c.Actions = append(c.Actions, checkAction)
	}
	if errInfo != nil {
		return errInfo
	}
	sleepAction := new(SleepPoint)
	sleepAction.SleepDuration = time.Duration(caseVersionOutput.SleepPoint)
	c.Actions = append(c.Actions, sleepAction)
	return
}
func graphqlChoiceVersion(ctx context.Context, method *siber.MethodInfo, version string) (query string, err error) {
	var afterGraphqlQuery []struct {
		afterVersion int32
		graphqlQuery string
	}
	tmpMethod := strings.Split(method.MethodName, ".")
	method.MethodName = tmpMethod[2]
	methodInfo, err := dao.NewDao().GetGraphqlQuery(ctx, method)
	if methodInfo == nil {
		return "", nil
	}
	if len(methodInfo.GraphqlQueryDetail) == 0 {
		return "", nil
	}
	for _, v := range methodInfo.GraphqlQueryDetail {
		afterVersion, err := api.RegexVersion(v.Version)
		if err != nil {
			return "", err
		}
		afterGraphqlQuery = append(afterGraphqlQuery, struct {
			afterVersion int32
			graphqlQuery string
		}{afterVersion: afterVersion, graphqlQuery: v.QueryString})
	}
	afterVersion, err := api.RegexVersion(version)
	if err != nil {
		return "", nil

	}
	for _, v := range afterGraphqlQuery {
		if v.afterVersion <= afterVersion {
			return v.graphqlQuery, nil
		}
	}
	return
}
func describeGraphQLRequest(ctx context.Context, c *Case, method *siber.MethodInfo) (err error) {
	methodTmp, err := dao.NewDao().SelectMethod(ctx, method)
	if err != nil {
		return
	}
	if methodTmp.MethodType == siberconst.HTTPProtocol || methodTmp.MethodType == siberconst.GRPCProtocol {
		return
	}

	strTmp := strings.Split(strings.Trim(methodTmp.GraphQuery, " "), " ")
	// query为空的，则不渲染这个子项
	if len(strTmp) == 0 {
		return
	}
	if len(strTmp) < 2 {
		graphQuery, err := graphqlChoiceVersion(ctx, method, c.Version)
		if err != nil {
			return err
		}
		if graphQuery == "" {
			return nil
		}
		methodTmp.GraphQuery = graphQuery
		strTmp = strings.Split(strings.Trim(methodTmp.GraphQuery, " "), " ")
		if len(strTmp) < 2 {
			err = status.Errorf(codes.InvalidArgument, "invalid graphQL query: %s", methodTmp.GraphQuery)
			return err
		}
	}
	opName := strings.Split(strTmp[1], "(")[0]
	methodTmp.GraphQuery = strings.ReplaceAll(methodTmp.GraphQuery, "\n", "\\n")
	c.Request.Body = []byte(fmt.Sprintf(`{"%s":"%s","%s":%s,"%s":"%s"}`,
		GraphQuery, methodTmp.GraphQuery,
		GraphVariables, string(c.Request.Body),
		GraphOperationName, opName))
	return
}
