package core

import (
	"api-test/api"
	"api-test/dao"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/globalsign/mgo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ManageProcessPlanInfo(ctx context.Context, ProcessPlanInput *siber.ManageProcessPlanInfo) (ProcessPlanOutput *siber.ProcessPlanInfo, err error) {
	// TODO: flow 格式合理性检查
	if ProcessPlanInput == nil {
		return
	}
	switch ProcessPlanInput.ManageMode {
	case api.CreateItemMode:
		ProcessPlanOutput, err = dao.NewDao().InsertProcessPlan(ctx, ProcessPlanInput.ProcessPlanInfo)
	case api.UpdateItemMode:
		ProcessPlanOutput, err = dao.NewDao().UpdateProcessPlan(ctx, ProcessPlanInput.ProcessPlanInfo)
	case api.QueryItemMode:
		ProcessPlanOutput, err = dao.NewDao().SelectProcessPlan(ctx, ProcessPlanInput.ProcessPlanInfo)
	case api.DeleteItemMode:
		ProcessPlanOutput, err = dao.NewDao().DeleteProcessPlan(ctx, ProcessPlanInput.ProcessPlanInfo)
	}
	return
}

func ManageProcessPlanList(ctx context.Context, request *siber.FilterInfo) (response *siber.ProcessPlanList, err error) {
	ProcessPlanList, totalNum, err := dao.NewDao().ListProcessPlan(ctx, request)
	if err != nil {
		return
	}
	response = &siber.ProcessPlanList{
		ProcessPlanInfo: *ProcessPlanList,
		TotalNum:        uint32(totalNum),
	}
	return
}

func ManageProcessPlanLog(ctx context.Context, request *siber.ProcessPlanLogRequest) (response *siber.ProcessPlanLogList, err error) {
	ProcessPlanLogList, totalNum, err := dao.NewDao().SelectProcessPlanLog(ctx, request)
	if err != nil {
		return
	}
	response = &siber.ProcessPlanLogList{
		LogInfo:  *ProcessPlanLogList,
		TotalNum: uint32(totalNum),
	}
	return
}

func ProcessRun(ctx context.Context, request *siber.ProcessRelease) (response *siber.ProcessResult, err error) {
	if request == nil || request.ProcessName == "" || request.EnvName == "" || request.Tag == "" {
		err = status.Errorf(codes.InvalidArgument, "Invalid Argument Error")
		return
	}
	processPlanInfo, err := dao.NewDao().SelectProcessPlan(ctx, &siber.ProcessPlanInfo{
		ProcessName: request.ProcessName,
	})
	if err == mgo.ErrNotFound {
		response = &siber.ProcessResult{
			ProcessName: request.ProcessName,
			Tag:         request.Tag,
			TestPass:    true,
		}
		return response, nil
	}
	if err != nil || processPlanInfo == nil {
		return
	}
	go func() {
		var planLogListInfo []*siber.PlanLog
		for _, p := range processPlanInfo.PlanInfo {
			runRequest := &siber.RunPlanRequest{
				PlanInfo: &siber.PlanInfo{
					PlanId: p.PlanId,
				},
				TriggerCondition: &siber.TriggerCondition{
					EnvironmentName: api.EnvironmentStage,
				},
			}
			p, err := DescribePlan(ctx, runRequest, CiTrigger)
			if err != nil {
				continue
			}
			planLog, err := p.Run(ctx)
			if planLog != nil {
				planLogListInfo = append(planLogListInfo, &siber.PlanLog{
					PlanLogId:  planLog.PlanLogId,
				})
			}
		}
		err = dao.NewDao().InsertProcessPlanLog(ctx, &siber.ProcessPlanLogInfo{
			ProcessName: request.ProcessName,
			Tag:         request.Tag,
			PlanLog:     planLogListInfo,
		})
		if err != nil {
			xzap.Logger(ctx).Error("Insert process plan logs error", zap.Any("err:", err))
			return
		}
	}()
	response = &siber.ProcessResult{
		ProcessName: request.ProcessName,
		Tag:         request.Tag,
		TestPass:    false,
	}
	return
}
func ProcessRunCheck(ctx context.Context, request *siber.ProcessRelease) (response *siber.ProcessResult, err error) {
	var logDetail *siber.PlanLog
	if request == nil || request.ProcessName == "" || request.EnvName == "" || request.Tag == "" {
		err = status.Errorf(codes.InvalidArgument, "Invalid Argument Error")
		return
	}
	_, err = dao.NewDao().SelectProcessPlan(ctx, &siber.ProcessPlanInfo{
		ProcessName: request.ProcessName,
	})
	if err == mgo.ErrNotFound {
		response = &siber.ProcessResult{
			ProcessName: request.ProcessName,
			Tag:         request.Tag,
			TestPass:    true,
			TestMsg:     "ProcessPlan NotFound",
		}
		return response, nil
	}
	processLog := &siber.ProcessPlanLogRequest{
		ProcessName: request.ProcessName,
		Tag:         request.Tag,
	}
	processPlanLog, _, err := dao.NewDao().SelectProcessPlanLog(ctx, processLog)
	if err != nil {
		xzap.Logger(ctx).Error("Select plan process plan logs error", zap.Any("err:", err))
		return
	}
	if processPlanLog == nil {
		response = &siber.ProcessResult{
			ProcessName: request.ProcessName,
			Tag:         request.Tag,
			TestPass:    false,
			TestMsg:     "ProcessPlan No Check",
		}
		return response, nil
	}
	response = &siber.ProcessResult{
		Tag:         request.Tag,
		ProcessName: request.ProcessName,
		EnvName:     request.EnvName,
		TestPass:    true,
	}
	for _, m := range *processPlanLog {
		for _, n := range m.PlanLog {
			logDetail, err = dao.NewDao().SelectPlanLogDetail(ctx, &siber.PlanLog{
				PlanLogId: n.PlanLogId,
			})
			if err != nil {
				xzap.Logger(ctx).Error("Select process plan log detail error", zap.Any("err:", err))
				return
			}
			if logDetail == nil {
				err = status.Errorf(codes.Aborted, "Select process plan log detail is nil")
				return
			}
			planStatus := logDetail.PlanStatus
			if planStatus == RunFailed {
				response.TestPass = false
				msg := "测试执行失败，详细信息：http://siber.wul.ai/plan/log?planId=" + n.PlanId
				response.TestMsg = msg
				return
			}
			if planStatus == Running {
				response.TestPass = false
				msg := "测试执行中，详细信息：http://siber.wul.ai/plan/log?planId=" + n.PlanId
				response.TestMsg = msg
				return
			}
		}
	}
	return
}
