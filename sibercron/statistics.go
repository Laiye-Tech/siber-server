package sibercron

import (
	"api-test/statistics"
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"time"
)

func CaseLogNumCron(ctx context.Context) (err error) {
	now := time.Now()
	lastDay := now.AddDate(0, 0, -30).Unix()
	_, err = statistics.CaseLogStat(ctx, &siber.GetCaseLogStatRequest{StartTime: uint64(lastDay), EndTime: uint64(time.Now().Unix())})
	return
}

func PlanLogNumCron(ctx context.Context) (err error) {
	now := time.Now()
	lastDay := now.AddDate(0, 0, -30).Unix()
	_, err = statistics.StatPlanLogNum(ctx, &siber.GetPlanLogStatRequest{StartTime: uint64(lastDay), EndTime: uint64(time.Now().Unix())})
	return
}