/**
* @Author: TongTongLiu
* @Date: 2020/4/20 4:57 下午
**/

package service

import (
	"api-test/api"
	"api-test/core"
	"api-test/log"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"go.uber.org/zap"
)

func (s *SiberService) ListCaseLog(ctx context.Context, request *siber.CaseLog) (response *siber.CaseLogList, err error) {
	xzap.Logger(ctx).Info("Case Log detail get request: %v", zap.Any("request", request))
	response, err = log.CaseLogList(ctx, request)
	return
}

func (s *SiberService) ListFlowLog(ctx context.Context, request *siber.FlowLog) (response *siber.FlowLogList, err error) {
	xzap.Logger(ctx).Info("Case Log detail get request: %v", zap.Any("request", request))
	response, err = log.FlowLogList(ctx, request)
	return
}

func (s *SiberService) ListPlanLog(ctx context.Context, request *siber.ListPlanLogRequest) (response *siber.PlanLogList, err error) {
	xzap.Logger(ctx).Info("Case Log detail get request: %v", zap.Any("request", request))
	response, err = log.PlanLogList(ctx, request)
	return
}

func (s *SiberService) ManageEnvList(ctx context.Context, request *siber.FilterInfo) (response *siber.EnvList, err error) {
	xzap.Logger(ctx).Info("manage Env list get request: %v ", zap.Any("request", request))
	response, err = core.ManageEnvList(ctx, request)
	return
}

func (s *SiberService) ManageTagList(ctx context.Context, request *siber.FilterInfo) (response *siber.TagList, err error) {
	xzap.Logger(ctx).Info("manage Tag list get request: %v ", zap.Any("request", request))
	response, err = core.ManageTagList(ctx, request)
	return
}
func (s *SiberService) ManagePlanList(ctx context.Context, request *siber.FilterInfo) (response *siber.PlanList, err error) {
	xzap.Logger(ctx).Info("manage plan list get request: %v ", zap.Any("request", request))
	response, err = core.ManagePlanList(ctx, request)
	return
}

func (s *SiberService) ManageFlowList(ctx context.Context, request *siber.FilterInfo) (response *siber.FlowList, err error) {
	xzap.Logger(ctx).Info("manage flow list get request: %v ", zap.Any("request", request))
	response, err = core.ManageFlowList(ctx, request)
	return
}

func (s *SiberService) ManageCaseList(ctx context.Context, request *siber.FilterInfo) (response *siber.CaseList, err error) {
	xzap.Logger(ctx).Info("manage case list get request: %v ", zap.Any("request", request))
	response, err = core.ManageCaseList(ctx, request)
	return
}

func (s *SiberService) ManageMethodList(ctx context.Context, request *siber.FilterInfo) (response *siber.MethodList, err error) {
	xzap.Logger(ctx).Info("manage method list get request: %v", zap.Any("request", request))
	response, err = api.GetMethodList(ctx, request)
	return
}

func (s *SiberService) ManageServiceList(ctx context.Context, request *siber.FilterInfo) (response *siber.ServiceList, err error) {
	xzap.Logger(ctx).Info("manage method list get request: %v", zap.Any("request", request))
	response, err = api.GetServiceList(ctx, request)
	return
}

func (s *SiberService) ManagePackageList(ctx context.Context, request *siber.FilterInfo) (response *siber.PackageList, err error) {
	xzap.Logger(ctx).Info("manage method list get request: %v", zap.Any("request", request))
	response, err = api.GetPackageList(ctx, request)
	return
}
func (s *SiberService) ManageProcessList(ctx context.Context, request *siber.ProcessListRequest) (response *siber.ProcessListResponse, err error) {
	xzap.Logger(ctx).Info("process list get request: %v", zap.Any("request", request))
	response, err = api.GetProcessList(request)
	return
}

func (s *SiberService) ManageProcessPlanList(ctx context.Context, request *siber.FilterInfo) (response *siber.ProcessPlanList, err error) {
	xzap.Logger(ctx).Info("process plan list get request: %v", zap.Any("request", request))
	response, err = core.ManageProcessPlanList(ctx, request)
	return
}

func (s *SiberService) ManageProcessPlanLog(ctx context.Context, request *siber.ProcessPlanLogRequest) (response *siber.ProcessPlanLogList, err error) {
	xzap.Logger(ctx).Info("process plan list get request: %v", zap.Any("request", request))
	response, err = core.ManageProcessPlanLog(ctx, request)
	return
}
