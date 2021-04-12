package core

import (
	"api-test/api"
	"api-test/dao"
	"api-test/payload"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Flow struct {
	Id            string
	Name          string
	Plan          *Plan
	Items         []*FlowItem
	RunMode       siber.RunModeType
	Variable      *payload.Variable
	BeforeActions []*FlowActionItem
	AfterActions  []*FlowActionItem

	FlowLog *siber.FlowLog
}

type FlowItem struct {
	Id    string
	Order int
	Case  *Case
}

func CreateFlow() *Flow {
	return &Flow{
		Variable: &payload.Variable{},
	}
}

func (f *Flow) LinkItem(items []*FlowItem) []*FlowItem {
	return nil
}

func (f *Flow) UnlinkItem(itemIds []int) error {
	return nil
}

type FlowActionItem struct {
	Id         int
	Order      int
	FlowAction FlowAction
}

func (f *Flow) LinkBeforeAction(items []*FlowActionItem) []*FlowActionItem {
	return nil
}

func (f *Flow) UnLinkBeforeAction(itemIds []int) error {
	return nil
}

func (f *Flow) LinkAfterAction(items []*FlowActionItem) []*FlowActionItem {
	return nil
}

func (f *Flow) UnLinkAfterAction(itemIds []int) error {
	return nil
}

func describeFlow(ctx context.Context, flow *Flow) (err error) {
	f := &siber.FlowInfo{
		FlowId:   flow.Id,
		FlowName: flow.Name,
	}
	ff, err := dao.NewDao().SelectFlow(ctx, f)
	if err != nil {
		return
	}
	i := 0
	for _, tmp := range ff.CaseList {
		flowItem := &FlowItem{
			Order: i,
			Case: &Case{
				Id: tmp,
			},
		}
		flow.Name = ff.FlowName
		flow.RunMode = ff.RunMode
		flowItem.Case.Plan = flow.Plan
		flowItem.Case.Flow = flow
		flow.Items = append(flow.Items, flowItem)
		i += 1
	}
	flow.Variable = &payload.Variable{}
	return
}

func (f *Flow) Run(ctx context.Context) (err error) {

	err = describeFlow(ctx, f)
	xzap.Logger(ctx).Info("run flow start", zap.Any("flow:", f.Name))
	if err != nil {
		return
	}
	for _, action := range f.BeforeActions {
		err = action.FlowAction.Execute()
		if err != nil {
			return
		}
	}
	f.Variable.Create()
	err = f.InitFlowLog(ctx)
	if err != nil {
		return
	}
	for _, item := range f.Items {
		c := item.Case
		errCase := c.Run(ctx)
		if errCase != nil {
			err = errCase
			c.finished(ctx, err)
			f.finished(ctx, err)
			if f.RunMode == siber.RunModeType_DEFAULT || f.RunMode == siber.RunModeType_ABORT {
				break
			}
			if f.RunMode == siber.RunModeType_IGNORE_ERROR {
				continue
			}
		}
	}

	for _, action := range f.AfterActions {
		err = action.FlowAction.Execute()
		if err != nil {
			return
		}
	}
	f.finished(ctx, err)
	xzap.Logger(ctx).Info("run flow finish", zap.Any("flow:", f.Name))
	return
}

func (f *Flow) finished(ctx context.Context, err error) {
	if err == nil {
		f.FlowLog.FlowStatus = RunSuccess
	} else {
		f.FlowLog.ErrContent = err.Error()
		f.FlowLog.FlowStatus = RunFailed
	}
	_, err = dao.NewDao().InsertFlowLog(ctx, f.FlowLog)
}

/*
* 维护plan：CREATE：创建，UPDATE：修改
 */
func ManageFlowInfo(ctx context.Context, flowInput *siber.ManageFlowInfo) (flowOutput *siber.FlowInfo, err error) {
	// TODO: flow 格式合理性检查
	if flowInput == nil {
		return
	}
	switch flowInput.ManageMode {
	case api.CreateItemMode:
		flowOutput, err = dao.NewDao().InsertFlow(ctx, flowInput.FlowInfo)
		return
	case api.UpdateItemMode:
		flowOutput, err = dao.NewDao().UpdateFlow(ctx, flowInput.FlowInfo)
	case api.QueryItemMode:
		flowOutput, err = dao.NewDao().SelectFlow(ctx, flowInput.FlowInfo)
	case api.DeleteItemMode:
		flowOutput, err = dao.NewDao().DeleteFlow(ctx, flowInput.FlowInfo)
		return
	}
	if err != nil || flowOutput == nil || flowOutput.CaseList == nil {
		return
	}
	var caseDetail *[]*siber.ManageFlowCaseSub
	caseDetail, err = dao.NewDao().SelectCaseByIDList(ctx, &flowOutput.CaseList)
	if err != nil {
		return
	}
	flowOutput.CaseListDetail = *caseDetail
	return
}

func ManageFlowList(ctx context.Context, request *siber.FilterInfo) (response *siber.FlowList, err error) {
	flowList, totalNum, err := dao.NewDao().ListFlow(ctx, request)
	if err != nil {
		return
	}
	response = &siber.FlowList{
		FlowInfoList: *flowList,
		TotalNum:     uint32(totalNum),
	}
	return
}
