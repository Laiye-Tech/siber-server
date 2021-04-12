/**
* @Author: TongTongLiu
* @Date: 2019-09-11 17:38
**/

package service

import (
	"context"
	"encoding/json"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"

	"api-test/api"
	"api-test/core"
	"api-test/log"
	"api-test/statistics"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

var iterationCache = cache.New(8*time.Hour, 60*time.Minute)

type SiberService struct {
}

func (s *SiberService) GetCaseLogStat(ctx context.Context, req *siber.GetCaseLogStatRequest) (*siber.GetCaseLogStatResponse, error) {
	stat, err := statistics.CaseLogStat(ctx, req)
	if err != nil {
		return nil, err
	}
	var list []*siber.CaseLogStat
	for _, v := range stat {
		caseTime, _ := strconv.Atoi(v.Time)
		list = append(list, &siber.CaseLogStat{
			TotalRunNum:      uint32(v.TotalRunNum),
			SuccessfulRunNum: uint32(v.SuccessfulRunNum),
			FailedRunNum:     uint32(v.FailedRunNum),
			Time:             uint32(caseTime),
		})
	}
	return &siber.GetCaseLogStatResponse{
		List: list,
	}, nil
}

func (s *SiberService) GetCaseStat(ctx context.Context, req *siber.GetCaseStatRequest) (*siber.GetCaseStatResponse, error) {
	stat, err := statistics.CaseStat(ctx, req)
	if err != nil {
		return nil, err
	}
	var list []*siber.CaseStat
	for _, v := range stat {
		caseTime, _ := strconv.Atoi(v.Time)
		list = append(list, &siber.CaseStat{
			TotalNum:    uint32(v.TotalNum),
			IncreaseNum: uint32(v.IncreaseNum),
			Time:        uint32(caseTime),
		})
	}
	return &siber.GetCaseStatResponse{
		List: list,
	}, nil
}

func (s *SiberService) GetPlanLogStat(ctx context.Context, req *siber.GetPlanLogStatRequest) (*siber.GetPlanLogStatResponse, error) {
	stat, err := statistics.StatPlanLogNum(ctx, req)
	if err != nil {
		return nil, err
	}
	var list []*siber.PlanLogStat
	for _, v := range stat {
		planTime, _ := strconv.Atoi(v.Time)
		list = append(list, &siber.PlanLogStat{
			TotalRunNum:      uint32(v.TotalRunNum),
			SuccessfulRunNum: uint32(v.SuccessfulRunNum),
			Time:             uint64(planTime),
		})
	}
	return &siber.GetPlanLogStatResponse{
		List: list,
	}, nil
}

func NewSiberService() *SiberService {
	return &SiberService{}
}

func (s *SiberService) ParseMethodList(ctx context.Context, request *siber.ParseMethodListRequest) (response *siber.ServiceInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("ParseMethodList get request", zap.Any("request", b))
	response, err = api.ParseMethodsFromProtoFile(ctx, request.ProtoFile, request.ServiceName)
	return response, err
}

func (s *SiberService) ManageCaseVersion(ctx context.Context, request *siber.ManageCaseVersionInfo) (response *siber.CaseVersionInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage case get request", zap.Any("requset", b))
	_, err = json.Marshal(request)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "ManageCaseVersion marshal request failed, err:%v", err)
		return
	}
	response, err = core.ManageCaseVersion(ctx, request)
	return
}

func (s *SiberService) ManageCase(ctx context.Context, request *siber.ManageCaseInfo) (response *siber.CaseInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage case get request : ", zap.Any("request", b))
	response, err = core.ManageCaseInfo(ctx, request)
	return
}

func (s *SiberService) ManageFlow(ctx context.Context, request *siber.ManageFlowInfo) (response *siber.FlowInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage flow get request :", zap.Any("request", b))
	response, err = core.ManageFlowInfo(ctx, request)
	return
}

func (s *SiberService) ManagePlan(ctx context.Context, request *siber.ManagePlanInfo) (response *siber.PlanInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage plan get request : %v", zap.Any("request", b))
	response, err = core.ManagePlanInfo(ctx, request)
	return
}

func (s *SiberService) ManageMethod(ctx context.Context, request *siber.ManageMethodInfo) (response *siber.MethodInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage method get request :%v", zap.Any("request", b))
	response, err = api.ManageMethodInfo(ctx, request)
	return
}

func (s *SiberService) ManageEnv(ctx context.Context, request *siber.ManageEnvInfo) (response *siber.EnvInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage env get request :%v", zap.Any("request", b))
	response, err = core.ManageEnvInfo(ctx, request)
	return
}
func (s *SiberService) ManageTag(ctx context.Context, request *siber.ManageTagInfo) (response *siber.TagInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("manage tag get request :%v", zap.Any("request", b))
	response, err = core.ManageTagInfo(ctx, request)
	return
}

func (s *SiberService) DescribeMethod(ctx context.Context, request *siber.MethodInfo) (response *siber.MethodDescribe, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("DescribeMethod get request: %v ", zap.Any("request", b))
	response, err = api.DescribeMethodFromProtoFile(ctx, request)
	return
}

func (s *SiberService) RunCase(ctx context.Context, request *siber.RunCaseRequest) (response *siber.CaseLog, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("run case get request: %v", zap.Any("request", b))
	response, err = core.RunCase(ctx, request)
	return
}



func (s *SiberService) RunPlan(ctx context.Context, request *siber.RunPlanRequest) (response *siber.PlanLog, err error) {
	if request == nil || request.PlanInfo == nil || request.TriggerCondition == nil {
		xzap.Logger(ctx).Error("run plan failed", zap.String("err", "Lack of necessary conditions"))
		err = status.Errorf(codes.Aborted, "Lack of necessary conditions")
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("run plan get request: %v", zap.String("request", string(b)))
	p, err := core.DescribePlan(ctx, request, core.ManualTrigger)
	if err != nil {
		return
	}
	response, err = p.Run(ctx)

	return
}

func (s *SiberService) CaseLogDetail(ctx context.Context, request *siber.CaseLog) (response *siber.CaseLog, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("Case Log detail get request: %v", zap.String("request", string(b)))
	response, err = log.CaseLogDetail(ctx, request)
	return
}

func (s *SiberService) ListProtoFile(ctx context.Context, request *siber.ProtoFileListRequests) (response *siber.ProtoFileListResponse, err error) {
	xzap.Logger(ctx).Info("start get proto files")
	response = api.GetProtoFileList(ctx, request)
	return
}
func (s *SiberService) ListGraphqlMethod(ctx context.Context, request *siber.GraphqlMethodListRequest) (response *siber.GraphqlMethodListResponse, err error) {
	xzap.Logger(ctx).Info("start get graphql files")
	response, err = api.GetGraphqlMethods(ctx, request)
	return
}

func (s *SiberService) GetGraphqlQuery(ctx context.Context, request *siber.MethodInfo) (response *siber.MethodInfo, err error) {
	xzap.Logger(ctx).Info("start get graphql files")
	response, err = api.GetGraphqlQuery(ctx, request)
	return
}

type TapdRes struct {
	Status int32
	Data   []DataItem
	Info   string
}

type DataItem struct {
	Iteration Iteration
}

type Iteration struct {
	Id          string
	Name        string
	WorkspaceId string
	Startdate   string
	Enddate     string
	Status      string
	Creator     string
	Created     string
	Modified    string
	ReleaseId   string
	Description string
}

func (s *SiberService) CIIterations(ctx context.Context, request *siber.GetIterationsRequest) (response *siber.GetIterationsResponse, err error) {
	response, err = api.CIIterations(ctx, request)
	if err != nil {
		xzap.Logger(ctx).Error(" CIIterations request failed :%v", zap.Any("err", err))
	}
	return
}

func (s *SiberService) ManageProcessPlan(ctx context.Context, request *siber.ManageProcessPlanInfo) (response *siber.ProcessPlanInfo, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("process list get request: %v", zap.Any("request", b))
	response, err = core.ManageProcessPlanInfo(ctx, request)
	return
}

func (s *SiberService) DescribeRequest(ctx context.Context, request *siber.MethodInfo) (response *siber.MethodDescribe, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	if err != nil {
		xzap.Logger(ctx).Error(" DescribeRequest Marshal request failed :%v", zap.Any("err", err))
	}
	xzap.Logger(ctx).Info("DescribeRequest get request:%v", zap.Any("request", b))
	response, err = api.DescribeRequest(ctx, request)
	return
}

func (s *SiberService) GetNodeInfo(ctx context.Context, request *siber.NodeInfoRequest) (response *siber.NodeInfoResponse, err error) {
	if request == nil {
		return
	}
	b, err := json.Marshal(request)
	xzap.Logger(ctx).Info("GetNodeInfo get request: %v", zap.Any("request", b))
	response, err = core.GetNodeInfo(ctx, request)
	return
}

func (s *SiberService) GetProcessResult(ctx context.Context, request *siber.ProcessRelease) (response *siber.ProcessResult, err error) {
	if request == nil {
		xzap.Logger(ctx).Warn("GetProcessResult get nil request")
		err = status.Errorf(codes.InvalidArgument, "GetProcessResult get nil request")
		return
	}
	//if !request.VersionIteration {
	//	xzap.Logger(ctx).Info("GetProcessResult has no iteration")
	//	response = &siber.ProcessResult{
	//		ProcessName: request.ProcessName,
	//		Tag:         request.Tag,
	//		TestPass:    true,
	//		TestMsg:     "GetProcessResult has no iteration",
	//	}
	//	return response, nil
	//}
	b, err := json.Marshal(request)
	if err != nil {
		xzap.Logger(ctx).Error("GetProcessResult marshal failed", zap.Any("err", err))
		return
	}
	xzap.Logger(ctx).Info("GetProcessResult get request", zap.Any("request", string(b)))

	switch request.EnvName {
	// 开发和其他不做操作
	case api.EnvironmentDev:
		response, err = core.CIRun(ctx, request, false)
	// 测试和灰度运行case
	case api.EnvironmentTest:
		response, err = core.CIRun(ctx, request, true)
	// 灰度按照强制列表执行
	case api.EnvironmentStage:
		response, err = core.ProcessRun(ctx, request)
		_, err = core.CIRun(ctx, request, true)
	// 生产返回灰度的运行结果
	case api.EnvironmentProd:
		response, err = core.ProcessRunCheck(ctx, request)
		_, err = core.CIRun(ctx, request, true)
	default:
		return
	}

	return
}
