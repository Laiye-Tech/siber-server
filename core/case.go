package core

import (
	"api-test/api"
	"api-test/dao"
	"api-test/payload"
	"context"
	"encoding/json"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

const (
	InstanceCase  = "instance"
	InterfaceCase = "interface"
	InjectCase    = "inject"
)

type Case struct {
	Id       string
	Name     string
	Version  string
	CaseMode string
	Plan     *Plan
	Flow     *Flow

	Method   *api.Method
	Instance api.Instance
	Request  *payload.Request
	Response *payload.Response
	Actions  []CaseAction

	CaseLog *siber.CaseLog
}

// 初始化 case 日志：代表 case 已进入处理阶段
func (c *Case) initCaseLog(ctx context.Context) (err error) {
	c.CaseLog = &siber.CaseLog{
		PlanLogId:    c.Plan.PlanLog.PlanLogId,
		FlowLogId:    c.Flow.FlowLog.FlowLogId,
		PlanName:     c.Plan.Name,
		PlanId:       c.Plan.Id,
		FlowName:     c.Flow.Name,
		FlowId:       c.Flow.Id,
		CaseId:       c.Id,
		CaseStatus:   Running,
		DbInsertTime: time.Now().UnixNano() / int64(time.Millisecond),
	}
	var resp *siber.CaseLog
	resp, err = dao.NewDao().UpsertCaseLog(ctx, c.CaseLog)
	c.CaseLog.CaseLogId = resp.CaseLogId
	return
}

// 记录请求模板，用户登记的原始内容
func (c *Case) templateCaseLog(ctx context.Context) (err error) {
	c.CaseLog.CaseName = c.Name
	c.CaseLog.DbUpdateTime = time.Now().UnixNano() / int64(time.Millisecond)
	c.CaseLog.VersionControl = c.Version
	c.CaseLog.RequestTemplate = new(siber.ResponseDetail)
	switch c.CaseMode {
	case InterfaceCase:
		c.CaseLog.MethodName = c.Method.Name
		trigger := c.Method.Interfaces[c.Plan.Trigger.Protocol]
		c.CaseLog.RequestTemplate.UrlParameter = trigger.Url()
	case InjectCase, InstanceCase:
	default:
		errMsg := "不支持的 case 类型：" + c.CaseMode
		xzap.Logger(ctx).Info(errMsg)
		err = status.Errorf(codes.InvalidArgument, errMsg)
		return
	}

	//深拷贝，避免渲染时，值被覆盖
	c.CaseLog.RequestTemplate.Header = make(map[string]string)
	for k, v := range c.Request.Header {
		c.CaseLog.RequestTemplate.Header[k] = v
	}

	c.CaseLog.RequestTemplate.Body = string(c.Request.Body)
	_, err = dao.NewDao().UpsertCaseLog(ctx, c.CaseLog)
	return
}

// 记录渲染后的内容，实际的请求内容
func (c *Case) renderCaseLog(ctx context.Context) (err error) {
	c.CaseLog.RequestValue = new(siber.ResponseDetail)
	// TODO: debug 可删
	bHeader, err := json.Marshal(c.Request.Header)
	if err != nil {
		return
	}
	xzap.Logger(ctx).Info("bHeader:" + string(bHeader))
	c.CaseLog.RequestValue.Header = c.Request.Header
	if c.CaseMode == InterfaceCase {
		trigger := c.Method.Interfaces[c.Plan.Trigger.Protocol]
		triggerUrl := trigger.Url()
		c.CaseLog.Url = triggerUrl
		c.CaseLog.RequestValue.UrlParameter = c.CaseLog.Url
	}

	c.CaseLog.RequestValue.Body = string(c.Request.Body)
	_, err = dao.NewDao().UpsertCaseLog(ctx, c.CaseLog)
	return
}

func (c *Case) Run(ctx context.Context) (err error) {

	err = c.initCaseLog(ctx)
	if err != nil {
		return
	}

	err = describeCase(ctx, c)
	if err != nil {
		return
	}
	err = c.templateCaseLog(ctx)
	if err != nil {
		return
	}
	err = c.Render(ctx, c.Flow.Variable)
	if err != nil {
		return err
	}
	// TODO: 临时debug 待删除
	bHeader, err := json.Marshal(c.Request.Header)
	if err != nil {
		return
	}
	xzap.Logger(ctx).Info("bHeader " + string(bHeader))
	err = c.renderCaseLog(ctx)
	if err != nil {
		return
	}
	// TODO: 应该可以根据描述出来的 case 直接 Invoke
	switch c.CaseMode {
	case InterfaceCase:
		// TODO: graphQL 的也不应该放在这儿
		err = describeGraphQLRequest(ctx, c, &siber.MethodInfo{
			MethodName: c.Method.Name,
		})
		if err != nil {
			return
		}
		// TODO: 此时就应该知道是还是 grpc、http了
		// TODO: 也不应该传环境，应该直接传url，Invoke 就是 Invoke
		// TODO: interface.Invoke
		c.Response, err = c.Method.Invoke(ctx, c.Plan.Trigger.Protocol, c.Request, c.Plan.Trigger.Environment)
	case InstanceCase:
		c.Response, err = c.Instance.Execute(ctx, c.Request)
	case InjectCase:

	default:
		err = status.Errorf(codes.InvalidArgument, "unsupported case type:%s", c.CaseMode)
	}
	if err != nil {
		return err
	}
	var errInfo error
	for _, action := range c.Actions {
		err = action.Execute(ctx, c)
		if err != nil {
			errInfo = err
		}
	}
	if errInfo != nil {
		return errInfo
	}
	c.finished(ctx, err)
	return
}

func (c *Case) finished(ctx context.Context, err error) {
	if c.CaseLog == nil {
		xzap.Logger(ctx).Info(`{"event":"customize_warn","key":"caseLog","value":100}cannot persistence nil case log`, zap.Any("err", err))
		err = status.Errorf(codes.InvalidArgument, "cannot persistence nil case log")
		return
	}
	if err == nil {
		c.CaseLog.CaseStatus = RunSuccess
	} else {
		c.CaseLog.ErrContent = err.Error()
		c.CaseLog.CaseStatus = RunFailed
	}
	err = c.PersistenceCaseLog(ctx)
}

/*
* 维护case信息：仅维护基础信息
 */
func ManageCaseInfo(ctx context.Context, caseInput *siber.ManageCaseInfo) (caseOutput *siber.CaseInfo, err error) {
	// TODO: case 格式合理性检查：1-循环依赖 2-依赖case是否存在指定变量
	if caseInput == nil || caseInput.CaseInfo == nil {
		xzap.Logger(ctx).Warn("ManageCaseInfo failed, case info is nil")
		err = status.Errorf(codes.InvalidArgument, "ManageCaseInfo failed, case info is nil")
		return
	}
	switch caseInput.ManageMode {
	case api.CreateItemMode:
		caseOutput, err = dao.NewDao().InsertCase(ctx, caseInput.CaseInfo)
	case api.UpdateItemMode:
		caseOutput, err = dao.NewDao().UpdateCase(ctx, caseInput.CaseInfo)
	case api.QueryItemMode:
		caseOutput, err = dao.NewDao().SelectCase(ctx, caseInput.CaseInfo)
		if err != nil {
			return
		}
		c := &Case{
			Id: caseInput.CaseInfo.CaseId,
		}
		caseList, err := sortCaseVersion(ctx, c)
		if status.Code(err) == codes.OutOfRange || len(caseList) == 0 {
			return caseOutput, nil
		} else if err != nil {
			return nil, err
		}
		caseVersionInput := &dao.CaseVersionStandard{
			CaseId:         caseInput.CaseInfo.CaseId,
			VersionControl: caseList[len(caseList)-1].CurrentVersion,
		}

		var caseVersionList []*siber.CaseVersionInfo
		for i, _ := range caseList {
			var caseVersion *siber.CaseVersionInfo
			if i == 0 {
				caseVersionInfo, err := dao.NewDao().SelectCaseVersion(ctx, caseVersionInput)
				if err != nil {
					return caseOutput, err
				}
				caseVersion, err = CaseVersionToProto(ctx, caseVersionInfo)
				if err != nil {
					return caseOutput, err
				}
			} else {
				caseVersion = &siber.CaseVersionInfo{
					CaseId:         c.Id,
					VersionControl: caseList[len(caseList)-1-i].CurrentVersion,
				}

			}
			caseVersionList = append(caseVersionList, caseVersion)
		}
		caseOutput.CaseVersion = caseVersionList
	case api.DeleteItemMode:
		caseOutput, err = dao.NewDao().DeleteCase(ctx, caseInput.CaseInfo)
	case api.DuplicateItemMode:
		caseInfo, err := dao.NewDao().SelectCase(ctx, caseInput.CaseInfo)
		if err != nil || caseInfo == nil {
			return nil, err
		}
		str := strconv.FormatInt(time.Now().Unix(), 10)
		caseInfo.CaseName = caseInfo.CaseName + "_" + str
		caseInfo, err = dao.NewDao().InsertCase(ctx, caseInfo)
		if err != nil {
			return nil, err
		}
		caseVersionInput := &dao.CaseVersionStandard{
			CaseId: caseInput.CaseInfo.CaseId,
		}
		caseVersionList, err := dao.NewDao().ListCaseVersionByID(ctx, caseVersionInput)
		if err != nil {
			return nil, err
		}
		for _, v := range *caseVersionList {
			caseVersionInfo, err := dao.NewDao().SelectCaseVersion(ctx, v)
			if err != nil {
				return nil, err
			}
			caseVersionInfo.CaseId = caseInfo.CaseId
			_, err = dao.NewDao().InsertCaseVersion(ctx, caseVersionInfo)
			if err != nil {
				return nil, err
			}
		}
		caseOutput = &siber.CaseInfo{
			CaseName: caseInfo.CaseName,
		}
	}
	return caseOutput, err
}

func ManageCaseList(ctx context.Context, request *siber.FilterInfo) (response *siber.CaseList, err error) {
	caseList, totalNum, err := dao.NewDao().ListCase(ctx, request)
	if err != nil {
		return
	}
	response = &siber.CaseList{
		CaseInfoList: *caseList,
		TotalNum:     uint32(totalNum),
	}
	return
}

// case 界面点击运行
func RunCase(ctx context.Context, request *siber.RunCaseRequest) (response *siber.CaseLog, err error) {
	if request == nil {
		return
	}
	timeStamp := int(time.Now().Unix())
	name := "auto_" + request.CaseName + "_" + strconv.Itoa(timeStamp)

	// 创建flow
	flowInfo, err := ManageFlowInfo(ctx, &siber.ManageFlowInfo{
		ManageMode: api.CreateItemMode,
		FlowInfo: &siber.FlowInfo{
			FlowName:  name,
			CaseList:  []string{request.CaseId},
			Automatic: 1,
		},
	})
	if err != nil || flowInfo == nil {
		return
	}

	// 如果是instance 类型case，给假的接口形式和环境。默认不允许存在只有对DB操作的plan

	// 创建plan
	planInfo, err := ManagePlanInfo(ctx, &siber.ManagePlanInfo{
		ManageMode: api.CreateItemMode,
		PlanInfo: &siber.PlanInfo{
			PlanName:       name,
			FlowList:       []string{flowInfo.FlowId},
			Automatic:      1,
			EnvironmentId:  request.EnvironmentId,
			InterfaceType:  request.InterfaceType,
			VersionControl: request.VersionControl,
		},
	})

	// 如果是mysql类型case，使用假的环境
	if err != nil || planInfo == nil {
		return
	}
	// 运行 plan
	runPlanRequest := &siber.RunPlanRequest{
		PlanInfo: planInfo,
		TriggerCondition: &siber.TriggerCondition{
			EnvironmentName: request.EnvironmentName,
		},
	}
	p, err := DescribePlan(ctx, runPlanRequest, ManualTrigger)
	if err != nil || p == nil {
		return
	}
	_, err = p.Run(ctx)
	if err != nil {
		return
	}
	caseLogFilter := &siber.CaseLog{
		PlanId: planInfo.PlanId,
		FlowId: flowInfo.FlowId,
		CaseId: request.CaseId,
	}
	// TODO: 这是一个不好的实现，需要优化
	var logs *[]*siber.CaseLog
	for i := 0; i < 20; i++ {
		logs, err = dao.NewDao().ListCaseLog(ctx, caseLogFilter)
		if err != nil {
			return
		}
		if logs == nil || len(*logs) == 0 || (*logs)[0].CaseStatus == Running {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if logs == nil || len(*logs) == 0 {
		err = status.Errorf(codes.Canceled, "超时，请在plan中重试")
		return
	}
	response, err = dao.NewDao().SelectCaseLogDetail(ctx, (*logs)[0])
	return
}
