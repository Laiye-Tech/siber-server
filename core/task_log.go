/**
* @Author: TongTongLiu
* @Date: 2019/10/15 4:29 下午
**/

package core

import (
	"context"
	"encoding/json"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"api-test/dao"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

const (
	RunInit    = 0
	Running    = 1
	RunSuccess = 2
	RunFailed  = 3
	RunAbort   = 9
)

const (
	InitLog   = "INIT"
	FinishLog = "FINISH"
)

// TODO: 这块儿写的不好，重构后删除
// TODO: instance inject 和 default 的还没有处理
//func (c *Case) InitCaseLog(ctx context.Context) (err error) {
//	if c == nil || c.Plan == nil || c.Plan.PlanLog == nil || c.Flow == nil || c.Flow.FlowLog == nil {
//		xzap.Logger(ctx).Warn("InitCaseLog failed, case or flow or plan info or log is nil")
//		err = status.Errorf(codes.FailedPrecondition, "plan or flow log is nil")
//		return
//	}
//	log := &siber.CaseLog{
//		PlanLogId:      c.Plan.PlanLog.PlanLogId,
//		FlowLogId:      c.Flow.FlowLog.FlowLogId,
//		PlanName:       c.Plan.Name,
//		PlanId:         c.Plan.Id,
//		FlowName:       c.Flow.Name,
//		FlowId:         c.Flow.Id,
//		CaseId:         c.Id,
//		CaseName:       c.Name,
//		CaseStatus:     Running,
//		DbInsertTime:   time.Now().UnixNano() / int64(time.Millisecond),
//		VersionControl: c.Version,
//	}
//	if c == nil || c.Plan.Trigger == nil || c.Plan.Trigger.Protocol == "" {
//		xzap.Logger(ctx).Warn("case is nil")
//		err = status.Errorf(codes.InvalidArgument, "case is nil")
//		return
//	}
//	switch c.CaseMode {
//	case InterfaceCase:
//		if c.Method == nil || c.Method.Interfaces == nil {
//			xzap.Logger(ctx).Warn("no valid interface")
//			err = status.Errorf(codes.FailedPrecondition, "no valid interface")
//			return
//		}
//		log.MethodName = c.Method.Name
//		trigger := c.Method.Interfaces[c.Plan.Trigger.Protocol]
//		log.Url = trigger.Url()
//
//	case InstanceCase:
//		if c.Instance == nil {
//			xzap.Logger(ctx).Warn("no valid instance")
//			err = status.Errorf(codes.FailedPrecondition, "no valid instance")
//			return
//		}
//		log.InstanceInfo, err = c.Instance.GetInfo(ctx)
//		if err != nil {
//			return
//		}
//	case InjectCase:
//	default:
//		err = status.Errorf(codes.FailedPrecondition, "unsupported case mode :%v", c.CaseMode)
//		return
//	}
//
//	var actions []string
//	for _, a := range c.Actions {
//		aByte, err := json.Marshal(a)
//		if err != nil {
//			xzap.Logger(ctx).Warn("marshal actions failed", zap.Any("action", c.Actions), zap.Any("err", err))
//			err = status.Errorf(codes.FailedPrecondition, "marshal action failed, action:%v, err:%v", c.Actions, err)
//			return err
//		}
//		actions = append(actions, string(aByte))
//	}
//
//	c.CaseLog = log
//	c.CaseLog.RequestTemplate = new(siber.ResponseDetail)
//	//深拷贝，避免渲染时，值被覆盖
//	if c.CaseMode == InterfaceCase {
//		trigger := c.Method.Interfaces[c.Plan.Trigger.Protocol]
//		triggerUrl := trigger.Url()
//		c.CaseLog.RequestTemplate.UrlParameter = triggerUrl
//	}
//	c.CaseLog.RequestTemplate.Header = make(map[string]string)
//	//TODO:修改为直接引用proto中的结构体，这里就不用一一赋值了
//	for k, v := range c.Request.Header {
//		c.CaseLog.RequestTemplate.Header[k] = v
//	}
//	//c.CaseLog.RequestTemplate.Header = c.Request.Header
//	c.CaseLog.RequestTemplate.Body = string(c.Request.Body)
//	_, err = dao.NewDao().InsertCaseLog(ctx, c.CaseLog)
//	return
//}

func (c *Case) PersistenceCaseLog(ctx context.Context) (err error) {
	//TODO:修改为直接引用proto中的结构体，这里就不用一一赋值了
	c.CaseLog.ResponseValue = new(siber.ResponseDetail)
	if c.Response != nil {
		c.CaseLog.ResponseValue.Header = c.Response.Header
		c.CaseLog.ResponseValue.Body = string(c.Response.Body)
		c.CaseLog.ResponseValue.StatusCode = int32(c.Response.StatusCode)
		c.CaseLog.ResponseValue.CostTime = int32(c.Response.TimeCost)
	}
	var actions []string
	for _, a := range c.Actions {
		aByte, err := json.Marshal(a)
		if err != nil {
			xzap.Logger(ctx).Warn("marshal actions failed", zap.Any("action", c.Actions), zap.Any("err", err))
			err = status.Errorf(codes.FailedPrecondition, "marshal action failed, action:%v, err:%v", c.Actions, err)
			return err
		}
		actions = append(actions, string(aByte))
	}
	c.CaseLog.ActionConsequence = actions
	_, err = dao.NewDao().UpsertCaseLog(ctx, c.CaseLog)
	return
}

func (f *Flow) InitFlowLog(ctx context.Context) (err error) {
	if f == nil || f.Plan == nil || f.Plan.PlanLog == nil {
		xzap.Logger(ctx).Warn("init flow log failed, get nil condition")
		err = status.Errorf(codes.FailedPrecondition, "init flow log failed, get nil condition")
		return
	}
	log := &siber.FlowLog{
		PlanLogId:    f.Plan.PlanLog.PlanLogId,
		PlanName:     f.Plan.Name,
		PlanId:       f.Plan.Id,
		FlowName:     f.Name,
		FlowId:       f.Id,
		FlowStatus:   Running,
		DbInsertTime: time.Now().Unix(),
	}
	f.FlowLog = log
	_, err = dao.NewDao().InsertFlowLog(ctx, f.FlowLog)
	return
}

func (p *Plan) InitPlanLog(ctx context.Context) (planLog *siber.PlanLog, err error) {
	if p == nil || p.Id == "" {
		xzap.Logger(ctx).Warn("init plan log failed, get nil condition")
		err = status.Errorf(codes.FailedPrecondition, "init plan log failed, get nil condition")
		return
	}
	log := &siber.PlanLog{
		PlanId:     p.Id,
		PlanName:   p.Name,
		PlanStatus: Running,
		//Trigger:              nil,
		EnvironmentName: p.Trigger.Environment,
		InterfaceType:   p.Trigger.Protocol,
		VersionControl:  p.Trigger.VersionControl,
		DbInsertTime:    time.Now().Unix(),
	}
	p.PlanLog = log
	planLog, err = dao.NewDao().InsertPlanLog(ctx, p.PlanLog)
	return
}
