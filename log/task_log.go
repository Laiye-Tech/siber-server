/**
* @Author: TongTongLiu
* @Date: 2019/10/16 10:15 上午
**/

package log

import (
	"api-test/dao"
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

func CaseLogDetail(ctx context.Context, request *siber.CaseLog) (log *siber.CaseLog, err error) {
	if request == nil {
		return
	}
	log, err = dao.NewDao().SelectCaseLogDetail(ctx, request)
	return
}

func CaseLogList(ctx context.Context, request *siber.CaseLog) (log *siber.CaseLogList, err error) {
	if request == nil {
		return
	}
	log = new(siber.CaseLogList)
	caseLogList, err := dao.NewDao().ListCaseLog(ctx, request)
	if err != nil {
		return
	}
	log.CaseLogList = *caseLogList
	return
}

func FlowLogList(ctx context.Context, request *siber.FlowLog) (log *siber.FlowLogList, err error) {
	if request == nil {
		return
	}
	log = new(siber.FlowLogList)
	flowLogList, err := dao.NewDao().ListFlowLog(ctx, request)
	if err != nil {
		return
	}
	log.FlowLogList = *flowLogList
	return
}

func PlanLogList(ctx context.Context, request *siber.ListPlanLogRequest) (log *siber.PlanLogList, err error) {
	if request == nil {
		return
	}
	log = new(siber.PlanLogList)
	planLogList, totalNum, err := dao.NewDao().ListPlanLog(ctx, request)
	if err != nil {
		return
	}
	log.PlanLogList = *planLogList
	log.TotalNum = uint32(totalNum)
	return
}
