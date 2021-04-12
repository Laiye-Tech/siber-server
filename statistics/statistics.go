package statistics

import (
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
	"strconv"
	"time"

	"github.com/astaxie/beego/cache"

	"api-test/dao"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

var statCache = cache.NewMemoryCache()

type StatCase struct {
	TotalNum    int
	IncreaseNum int
	Time        string
}

func CaseStat(ctx context.Context, request *siber.GetCaseStatRequest) ([]StatCase, error) {
	var statCase []StatCase
	key := "case" + strconv.Itoa(int(request.EndTime)) + strconv.Itoa(int(request.StartTime))
	if statCache.IsExist(key) {
		return statCache.Get(key).([]StatCase), nil
	}
	// 起始时间 到 第二天 0点
	var begin, end int64
	beginTime := time.Unix(int64(request.StartTime), 0)
	beinWeeHour := time.Date(beginTime.Year(), beginTime.Month(), beginTime.Day(), 0, 0, 0, 0, beginTime.Location())
	NextDayTimestamp := beinWeeHour.Unix() + 86400

	for end <= int64(request.EndTime)+86400 {
		if begin == 0 {
			begin = int64(request.StartTime)
		}
		if end == 0 {
			if NextDayTimestamp <= int64(request.EndTime) {
				end = NextDayTimestamp
			} else {
				end = int64(request.EndTime)
			}
		}

		increaseNumQuery := bson.M{
			"inserttime":  bson.M{"$lt": end, "$gt": begin},
			"invaliddate": 0,
		}
		increaseNum, err := dao.StatCase(ctx, increaseNumQuery)
		if err != nil {
			return statCase, err
		}
		totalNumQuery := bson.M{
			"inserttime":  bson.M{"$lt": end},
			"invaliddate": 0,
		}
		totalNum, err := dao.StatCase(ctx, totalNumQuery)
		if err != nil {
			return statCase, err
		}
		tm := time.Unix(begin, 0)
		day := strconv.Itoa(tm.Day())
		if tm.Day() < 10 {
			day = "0" + strconv.Itoa(tm.Day())
		}
		statCase = append(statCase, StatCase{
			TotalNum:    totalNum,
			IncreaseNum: increaseNum,
			Time:        strconv.Itoa(tm.Year()) + strconv.Itoa(int(tm.Month())) + day,
		})
		begin, end = end, end+86400
	}
	err := statCache.Put(key, statCase, time.Hour*24)
	if err != nil {
		xzap.Logger(ctx).Error("put into the cache error", zap.Any("err", err))
	}
	return statCase, nil
}

type StatCaseLog struct {
	TotalRunNum      int
	SuccessfulRunNum int
	FailedRunNum     int
	Time             string
}

func getCaseLog(ctx context.Context, begin int64, end int64) (StatCaseLog, error) {
	totalNumQuery := bson.M{
		"dbinserttime": bson.M{"$lt": end * 1000, "$gt": begin * 1000},
	}
	totalNum, err := dao.StatCaseLog(ctx, totalNumQuery)
	if err != nil {
		return StatCaseLog{}, err
	}

	successfulNumQuery := bson.M{
		"dbinserttime": bson.M{"$lt": end * 1000, "$gt": begin * 1000},
		"casestatus":   2,
	}
	successfulNum, err := dao.StatCaseLog(ctx, successfulNumQuery)
	if err != nil {
		return StatCaseLog{}, err
	}

	failedNumQuery := bson.M{
		"dbinserttime": bson.M{"$lt": end * 1000, "$gt": begin * 1000},
		"casestatus":   3,
	}
	failedNum, err := dao.StatCaseLog(ctx, failedNumQuery)
	if err != nil {
		return StatCaseLog{}, err
	}

	tm := time.Unix(begin, 0)
	day := strconv.Itoa(tm.Day())
	if tm.Day() < 10 {
		day = "0" + strconv.Itoa(tm.Day())
	}
	statCaseLog := StatCaseLog{
		TotalRunNum:      totalNum,
		SuccessfulRunNum: successfulNum,
		FailedRunNum:     failedNum,
		Time:             strconv.Itoa(tm.Year()) + strconv.Itoa(int(tm.Month())) + day,
	}
	return statCaseLog, err
}
func CaseLogStat(ctx context.Context, request *siber.GetCaseLogStatRequest) (CaseLog []StatCaseLog, err error) {
	var statCaseLog []StatCaseLog
	// 起始时间 到 第二天 0点
	var begin, end int64
	var getStatCaseLog StatCaseLog

	beginTime := time.Unix(int64(request.StartTime), 0)
	beginWeeHour := time.Date(beginTime.Year(), beginTime.Month(), beginTime.Day(), 0, 0, 0, 0, beginTime.Location())
	NextDayTimestamp := beginWeeHour.Unix() + 86400

	for end <= int64(request.EndTime)+86400 {
		if time.Unix(begin, 0).Format("20060102") == time.Unix(int64(request.EndTime), 0).Format("20060102") {
			break
		}
		if begin == 0 {
			begin = int64(request.StartTime)
		}
		if end == 0 {
			if NextDayTimestamp <= int64(request.EndTime) {
				end = NextDayTimestamp
			} else {
				end = int64(request.EndTime)
			}
		}
		key := "caselog" + strconv.Itoa(int(begin)) + strconv.Itoa(int(end))
		if statCache.IsExist(key) {
			getStatCaseLog = statCache.Get(key).(StatCaseLog)

		} else {
			getStatCaseLog, err = getCaseLog(ctx, begin, end)
			if err != nil {
				xzap.Logger(ctx).Error("get stat case log error", zap.Any("err", err))
			}
			err = statCache.Put(key, getStatCaseLog, time.Hour*24)
			if err != nil {
				xzap.Logger(ctx).Error("put into the cache error", zap.Any("err", err))
			}
		}
		statCaseLog = append(statCaseLog, getStatCaseLog)
		begin, end = end, end+86400
	}
	getStatCaseLog, err = getCaseLog(ctx, begin, end)
	if err != nil {
		xzap.Logger(ctx).Error("get stat case log error", zap.Any("err", err))
	}
	statCaseLog = append(statCaseLog, getStatCaseLog)
	return statCaseLog, nil
}

type StatPlanLog struct {
	TotalRunNum      int
	SuccessfulRunNum int
	Time             string
}

func getPlanLog(ctx context.Context, begin int64, end int64) (StatPlanLog, error) {
	totalNumQuery := bson.M{
		"dbinserttime": bson.M{"$lt": end, "$gt": begin},
	}
	totalNum, err := dao.StatPlan(ctx, totalNumQuery)
	if err != nil {
		return StatPlanLog{}, err
	}
	successfulNumQuery := bson.M{
		"dbinserttime": bson.M{"$lt": end, "$gt": begin},
		"planstatus":   2,
	}
	successfulNum, err := dao.StatPlan(ctx, successfulNumQuery)
	if err != nil {
		return StatPlanLog{}, err

	}
	tm := time.Unix(begin, 0)
	day := strconv.Itoa(tm.Day())
	if tm.Day() < 10 {
		day = "0" + strconv.Itoa(tm.Day())
	}

	getStatPlanLog := StatPlanLog{
		TotalRunNum:      totalNum,
		SuccessfulRunNum: successfulNum,
		Time:             strconv.Itoa(tm.Year()) + strconv.Itoa(int(tm.Month())) + day,
	}
	return getStatPlanLog, err
}
func StatPlanLogNum(ctx context.Context, request *siber.GetPlanLogStatRequest) (PlanLog []StatPlanLog, err error) {
	var statPlanLog []StatPlanLog
	var begin, end int64
	var getStatPlanLog StatPlanLog
	beginTime := time.Unix(int64(request.StartTime), 0)
	beginWeeHour := time.Date(beginTime.Year(), beginTime.Month(), beginTime.Day(), 0, 0, 0, 0, beginTime.Location())
	NextDayTimestamp := beginWeeHour.Unix() + 86400
	for end <= int64(request.EndTime)+86400 {
		if time.Unix(begin, 0).Format("20060102") == time.Unix(int64(request.EndTime), 0).Format("20060102") {
			break
		}
		if begin == 0 {
			begin = int64(request.StartTime)
		}
		if end == 0 {
			if NextDayTimestamp <= int64(request.EndTime) {
				end = NextDayTimestamp
			} else {
				end = int64(request.EndTime)
			}
		}

		key := "planlog" + strconv.Itoa(int(begin)) + strconv.Itoa(int(end))
		if statCache.IsExist(key) {
			getStatPlanLog = statCache.Get(key).(StatPlanLog)
		} else {
			getStatPlanLog, err = getPlanLog(ctx, begin, end)
			if err != nil {
				xzap.Logger(ctx).Error("get stat plan log error", zap.Any("err", err))
			}
			err = statCache.Put(key, getStatPlanLog, time.Hour*24)
			if err != nil {
				xzap.Logger(ctx).Error("put into the cache error", zap.Any("err", err))
			}
		}
		begin, end = end, end+86400
		statPlanLog = append(statPlanLog, getStatPlanLog)
	}
	getStatPlanLog, err = getPlanLog(ctx, begin, end)
	if err != nil {
		xzap.Logger(ctx).Error("get stat plan log error", zap.Any("err", err))
	}
	statPlanLog = append(statPlanLog, getStatPlanLog)

	return statPlanLog, nil
}
