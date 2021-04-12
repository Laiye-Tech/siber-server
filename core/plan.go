package core

import (
	"api-test/api"
	"api-test/configs"
	"api-test/dao"
	"api-test/payload"
	"api-test/sibercron"
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

const (
	PlanSymbol = "plan"
	FlowSymbol = "flow"
	CaseSymbol = "case"
)

const RoutineNum = 10 // 并发数

const (
	ManualTrigger = "手动触发"
	CiTrigger     = "ci触发"
	CronTrigger   = "定时任务触发"
)

var ScheduleManagerGlobal *ScheduleManager

func init() {
	ScheduleManagerGlobal = NewScheduleManager()
}

type Plan struct {
	Id    string
	Name  string
	Url   string
	Flows []*Flow

	Trigger *Trigger
	Threads int32
	PlanLog *siber.PlanLog
}

func createPlan() *Plan {
	return &Plan{}
}

func DescribePlan(ctx context.Context, p *siber.RunPlanRequest, triggerType string) (plan *Plan, err error) {
	planInfo, err := dao.NewDao().SelectPlan(ctx, p.PlanInfo)
	if err != nil {
		return
	}
	plan = createPlan()
	xzap.Logger(ctx).Info("run plan start", zap.Any("plan:", planInfo.PlanName), zap.Any("triggerType:", triggerType))
	plan.Name = planInfo.PlanName
	var flowItems []*Flow
	for _, f := range planInfo.FlowList {
		flow := &Flow{
			Id: f,
		}
		flow.Plan = plan
		flowItems = append(flowItems, flow)
	}

	url, err := api.GetProtocolURL(ctx, planInfo.InterfaceType, planInfo.EnvironmentId, p.TriggerCondition.EnvironmentName)
	if err != nil {
		return
	}

	trigger := &Trigger{
		TriggerType:    triggerType,
		Protocol:       planInfo.InterfaceType,
		Environment:    p.TriggerCondition.EnvironmentName,
		VersionControl: planInfo.VersionControl,
	}

	plan.Id = planInfo.PlanId
	plan.Url = url
	plan.Flows = flowItems
	plan.Trigger = trigger
	plan.Threads = planInfo.Threads
	return
}

func flowProducer(ctx context.Context, plan *Plan) <-chan *Flow {
	ch := make(chan *Flow, RoutineNum)
	go func() {
		for _, flows := range plan.Flows {
			ch <- flows
		}
		close(ch)
	}()
	return ch
}

func planSubConsumer(ctx context.Context, ch <-chan *Flow) (errFirst error) {
	for {
		flow, ok := <-ch
		if !ok {
			break
		}
		err := flow.Run(ctx)
		if err != nil {
			flow.finished(ctx, err)
			if errFirst == nil {
				errFirst = err
			}
			continue
		}
	}
	return
}

func flowConsumer(ctx context.Context, p *Plan, ch <-chan *Flow) (errPlan error) {
	wg := &sync.WaitGroup{}
	if p.Threads == 0 {
		p.Threads = RoutineNum
	}
	for r := 0; r < int(p.Threads); r++ {
		wg.Add(1)
		go func() {
			err := planSubConsumer(ctx, ch)
			if err != nil && errPlan == nil {
				errPlan = err
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return
}

func (p *Plan) Run(ctx context.Context) (planLogInfo *siber.PlanLog, err error) {
	// TODO:需要传入环境
	planLogInfo, err = p.InitPlanLog(ctx)
	if err != nil {
		return
	}
	go func() {
		ch := flowProducer(ctx, p)
		err = flowConsumer(ctx, p, ch)
		p.finished(ctx, err)
		xzap.Logger(ctx).Info("run plan finish", zap.Any("plan:", p.Name))
	}()
	return
}

func (p *Plan) finished(ctx context.Context, err error) () {
	if err == nil {
		p.PlanLog.PlanStatus = RunSuccess
	} else {
		p.PlanLog.ErrContent = err.Error()
		p.PlanLog.PlanStatus = RunFailed
		privateDeploy := configs.GetGlobalConfig().Flag.PrivateDeploy
		// 私有部署不发送消息
		if !privateDeploy {
			sendMsg := GeneratePlanMsg(p.Id, p.Name, p.Trigger, err)
			payload.SendPlanRes(ctx, sendMsg)
		}
	}
	p.PlanLog.Trigger = p.Trigger.TriggerType
	_, err = dao.NewDao().InsertPlanLog(ctx, p.PlanLog)
	return
}

/*
* 维护plan：CREATE：创建，UPDATE：修改
 */
func ManagePlanInfo(ctx context.Context, planInput *siber.ManagePlanInfo) (planOutput *siber.PlanInfo, err error) {
	if planInput == nil || planInput.PlanInfo == nil {
		return
	}
	// 校验格式
	err = planFormatVerify(ctx, planInput.PlanInfo)
	if err != nil {
		return
	}
	switch planInput.ManageMode {
	case api.CreateItemMode:
		err = sibercron.UpdatePlanServices(ctx, planInput.PlanInfo)
		if err != nil {
			return
		}
		planOutput, err = dao.NewDao().InsertPlan(ctx, planInput.PlanInfo)
		if err == nil {
			err = ScheduleManagerGlobal.EditCron(ctx, planOutput)
		}
	case api.UpdateItemMode:
		err = sibercron.UpdatePlanServices(ctx, planInput.PlanInfo)
		if err != nil {
			return
		}
		planOutput, err = dao.NewDao().UpdatePlan(ctx, planInput.PlanInfo)
		if err == nil && planOutput != nil {
			err = ScheduleManagerGlobal.EditCron(ctx, planOutput)
		}
	case api.QueryItemMode:
		planOutput, err = dao.NewDao().SelectPlan(ctx, planInput.PlanInfo)
		if err != nil {
			return
		}
	case api.DeleteItemMode:
		planOutput, err = dao.NewDao().DeletePlan(ctx, planInput.PlanInfo)
		if err == nil {
			err = ScheduleManagerGlobal.EditCron(ctx, planInput.PlanInfo)
		}
	}
	if err != nil || planOutput == nil {
		return
	}
	// 拼上环境信息，方便前端展示
	if planOutput.EnvironmentId != "" {
		EnvInfo := &siber.EnvInfo{
			EnvId: planOutput.EnvironmentId,
		}
		envInfoOutput, err := dao.NewDao().SelectEnv(ctx, EnvInfo)
		if err != nil {
			return nil, err
		}
		planOutput.EnvInfo = envInfoOutput
	}
	return planOutput, err
}

func ManagePlanList(ctx context.Context, request *siber.FilterInfo) (response *siber.PlanList, err error) {
	planList, totalNum, err := dao.NewDao().ListPlan(ctx, request)
	if err != nil {
		return
	}
	response = &siber.PlanList{
		PlanInfoList: *planList,
		TotalNum:     uint32(totalNum),
	}
	return
}

func GeneratePlanMsg(planId string, planName string, trigger *Trigger, errMsg error) string {
	var msg = "【集成测试平台自动测试通知】\n以下Plan未通过自动测试：\n"
	var innerMsg = fmt.Sprintf("\n PlanName: %s \n 触发方式: %s \n errMsg: %v \n time: %s \n", planName, trigger.TriggerType, errMsg, time.Now().String())
	msg += innerMsg
	msg += "详细信息：http://siber.wul.ai/plan/log?planId=" + planId
	return msg
}

func planFormatVerify(ctx context.Context, info *siber.PlanInfo) (err error) {
	if info == nil || info.TriggerCondition == nil || len(info.TriggerCondition) == 0 {
		return
	}

	// 校验定时任务是否符合预期
	for _, t := range info.TriggerCondition {
		if t.TriggerCron == "" {
			continue
		}
		c := cron.New(cron.WithSeconds())
		defer c.Stop()
		_, err = c.AddFunc(t.TriggerCron, nil)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "invalid cron format, %v", err)
			break
		}
	}
	return
}

func CIRun(ctx context.Context, request *siber.ProcessRelease, bind bool) (response *siber.ProcessResult, err error) {
	if request == nil || request.ProcessName == "" || request.EnvName == "" {
		return
	}
	// 寻找符合条件的plan
	go func() {
		var filter *siber.FilterInfo
		if bind == true {
			filter = &siber.FilterInfo{
				FilterContent: map[string]string{
					"bind_services":    request.ProcessName,
					"environment_name": request.EnvName,
				},
			}

		} else {
			filter = &siber.FilterInfo{
				FilterContent: map[string]string{
					"services": request.ProcessName,
				},
			}
		}
		planList, err := ManagePlanList(ctx, filter)

		if err != nil || planList == nil {
			return
		}
		for _, p := range planList.PlanInfoList {
			runRequest := &siber.RunPlanRequest{
				PlanInfo: &siber.PlanInfo{
					PlanId: p.PlanId,
				},
				TriggerCondition: &siber.TriggerCondition{
					EnvironmentName: request.EnvName,
				},
			}
			p, err := DescribePlan(ctx, runRequest, CiTrigger)
			if err != nil {
				continue
			}
			_, err = p.Run(ctx)
		}
	}()
	response = &siber.ProcessResult{
		ProcessName: request.ProcessName,
		Tag:         request.Tag,
		TestPass:    false,
	}
	return
}
