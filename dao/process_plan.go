package dao

import (
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/globalsign/mgo"
	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

type processPlanInfo struct {
	Id              string                `bson:"_id,omitempty"`
	ProcessPlanInfo siber.ProcessPlanInfo `bson:",inline"`
}

func (dao *Dao) SelectProcessPlan(ctx context.Context, processPlanInput *siber.ProcessPlanInfo) (processPlanOutput *siber.ProcessPlanInfo, err error) {
	err = isInputNull(ctx, processPlanInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionProcessPlan)
	if err != nil {
		return
	}
	processPlanOutput = new(siber.ProcessPlanInfo)
	if processPlanInput.ProcessPlanId != "" {
		err = c.Find(bson.M{"processplanid": processPlanInput.ProcessPlanId}).One(&processPlanOutput)
		if err != nil {
			return
		}
	} else {
		err = c.Find(bson.M{"processname": processPlanInput.ProcessName, "invaliddate": 0}).One(&processPlanOutput)
		if err != nil {
			return
		}
	}
	// 当存在ID时，根据ID进行查询

	return
}

func (dao *Dao) InsertProcessPlan(ctx context.Context, processPlanInput *siber.ProcessPlanInfo) (processPlanOutput *siber.ProcessPlanInfo, err error) {
	var planInfoOutput *siber.PlanInfo
	err = isInputNull(ctx, processPlanInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionProcessPlan)
	if err != nil {
		return
	}
	var planInfoList []*siber.PlanInfo
	if processPlanInput.PlanInfo == nil {
		err = status.Errorf(codes.Aborted, "plan info is nil")
		return
	}
	for _, m := range processPlanInput.PlanInfo {
		if m == nil {
			err = status.Errorf(codes.Aborted, "plan info is nil")
			return
		}
		planInfoOutput, err = dao.SelectPlan(ctx, &siber.PlanInfo{PlanId: m.PlanId})
		if err != nil {
			return
		}
		if planInfoOutput == nil {
			err = status.Errorf(codes.Aborted, "get plan info failed, planID: %s, err: %v", m.PlanId, err)
			return
		}
		planInfo := &siber.PlanInfo{
			PlanId: planInfoOutput.PlanId,
		}
		planInfoList = append(planInfoList, planInfo)
	}
	// 当存在ID时，根据ID进行查询

	mongoID := bson2.NewObjectId().Hex()
	processPlanInput.ProcessPlanId = mongoID
	processPlanInput.PlanInfo = planInfoList
	processPlanInfo := &processPlanInfo{
		Id:              mongoID,
		ProcessPlanInfo: *processPlanInput,
	}

	err = c.Insert(processPlanInfo)

	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "process name is a duplicate")
			return
		}

		err = status.Errorf(codes.Aborted, "process plan failed, err: %v", err)
		return
	}
	processPlanOutput, err = dao.SelectProcessPlan(ctx, processPlanInput)
	return
}

func (dao *Dao) DeleteProcessPlan(ctx context.Context, processPlanInput *siber.ProcessPlanInfo) (processPlanOutput *siber.ProcessPlanInfo, err error) {
	err = isInputNull(ctx, processPlanInput)
	if err != nil {
		return
	}
	if processPlanInput.ProcessPlanId == "" {
		err = status.Errorf(codes.InvalidArgument, "Process Plan Id is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionProcessPlan)
	if err != nil {
		return
	}
	selector := bson.M{"_id": processPlanInput.ProcessPlanId}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	processPlanOutput = new(siber.ProcessPlanInfo)
	return
}

func (dao *Dao) UpdateProcessPlan(ctx context.Context, processPlanInput *siber.ProcessPlanInfo) (processPlanOutput *siber.ProcessPlanInfo, err error) {
	var planInfoOutput *siber.PlanInfo
	err = isInputNull(ctx, processPlanInput)
	if err != nil {
		return
	}
	if processPlanInput.ProcessPlanId == "" {
		err = status.Errorf(codes.InvalidArgument, "Process Plan Id is null")
		return
	}
	var planInfoList []*siber.PlanInfo
	if processPlanInput.PlanInfo == nil {
		err = status.Errorf(codes.Aborted, "plan info is nil")
		return
	}
	for _, m := range processPlanInput.PlanInfo {
		if m == nil {
			err = status.Errorf(codes.Aborted, "plan info is nil")
			return
		}
		planInfoOutput, err = dao.SelectPlan(ctx, &siber.PlanInfo{PlanId: m.PlanId})
		if err != nil {
			return
		}
		if planInfoOutput == nil {
			err = status.Errorf(codes.Aborted, "get plan info failed, planID: %s, err: %v", m.PlanId, err)
			return
		}
		planInfo := &siber.PlanInfo{
			PlanId:   planInfoOutput.PlanId,
			PlanName: planInfoOutput.PlanName,
		}
		planInfoList = append(planInfoList, planInfo)
	}
	processPlanInput.PlanInfo = planInfoList
	processPlanInput.UpdateTime = time.Now().Unix()

	c, err := getMongoCollection(ctx, CollectionProcessPlan)
	if err != nil {
		return
	}
	err = c.Update(bson.M{"processplanid": processPlanInput.ProcessPlanId}, bson.M{"$set": processPlanInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "process name is a duplicate")
			return
		}
		err = status.Errorf(codes.Aborted, "update Process Plan failed, err: %v", err)
		return
	}
	processPlanOutput, err = dao.SelectProcessPlan(ctx, processPlanInput)
	return
}

func (dao *Dao) ListProcessPlan(ctx context.Context, filterInfo *siber.FilterInfo) (planList *[]*siber.ProcessPlanInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionProcessPlan)
	if err != nil {
		return
	}
	var query bson.M
	var ProcessPlans []*siber.ProcessPlanInfo
	var planInfoOutput *siber.PlanInfo
	page := defaultPage
	pageSize := defaultPageSize
	if _, ok := filterInfo.FilterContent["page"]; ok {
		page, _ = strconv.Atoi(filterInfo.FilterContent["page"])
		delete(filterInfo.FilterContent, "page")
	}
	if _, ok := filterInfo.FilterContent["page_size"]; ok {
		pageSize, _ = strconv.Atoi(filterInfo.FilterContent["page_size"])
		delete(filterInfo.FilterContent, "page_size")
	}

	// 分页查询 content传""
	if _, ok := filterInfo.FilterContent["content"]; ok {
		var filter = make(map[string]string)
		filter["processname"] = filterInfo.FilterContent["content"]
		var orQueries []bson.M
		for k, v := range filter {
			orQueries = append(orQueries, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		var andQuery []bson.M
		andQuery = append(andQuery, bson.M{"invaliddate": 0})
		andQuery = append(andQuery, bson.M{"$or": orQueries})
		query = bson.M{"$and": andQuery}
		err = c.Find(query).Sort("-updatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&ProcessPlans)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get process plan list failed, err : %v", err)
			return
		}
		totalNum, err = c.Find(query).Count()
		if err != nil {
			err = status.Errorf(codes.Aborted, "get process plan total num failed, err : %v", err)
			return
		}
	}
	if query == nil {
		var andQueries []bson.M
		andQueries = append(andQueries, bson.M{"invaliddate": 0})
		query = bson.M{"$and": andQueries}
	}
	err = c.Find(query).All(&ProcessPlans)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get process plan list failed, err : %v", err)
		return
	}
	for _, m := range ProcessPlans {
		for _, n := range m.PlanInfo {
			planInfoOutput, err = dao.SelectPlan(ctx, &siber.PlanInfo{PlanId: n.PlanId})
			if err != nil {
				return
			}
			if planInfoOutput == nil {
				err = status.Errorf(codes.Aborted, "get plan info failed, planID: %s, err: %v", planInfoOutput.PlanId, err)
				return
			}
			n.PlanName = planInfoOutput.PlanName
			n.PlanId = planInfoOutput.PlanId
		}
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan total num failed, err : %v", err)
		return
	}
	return &ProcessPlans, totalNum, err
}

// 查询日志
func (dao *Dao) SelectProcessPlanLog(ctx context.Context, processPlanLogInput *siber.ProcessPlanLogRequest) (processPlanLogOutput *[]*siber.ProcessPlanLogInfo, totalNum int, err error) {
	err = isInputNull(ctx, processPlanLogInput)

	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionProcessPlanLog)
	if err != nil {
		return
	}
	var logs []*siber.ProcessPlanLogInfo
	if processPlanLogInput.ProcessName == "" {
		err = status.Errorf(codes.Aborted, "get process plan log failed, process name is nil")
		return
	}
	if processPlanLogInput.Tag == "" {
		// 当存在ID时，根据ID进行查询
		query := bson.M{"processname": processPlanLogInput.ProcessName}
		err = c.Find(query).Sort("-updatetime").All(&logs)

		if err != nil {
			err = status.Errorf(codes.Aborted, "get process plan log info failed, process plan id: %s, err: %v", processPlanLogInput.ProcessName, err)
			return
		}
		totalNum, err = c.Find(query).Count()
		if err != nil {
			err = status.Errorf(codes.Aborted, "get process plan log total num failed, err : %v", err)
			return
		}
	}
	if processPlanLogInput.Tag != "" {
		query := bson.M{"processname": processPlanLogInput.ProcessName, "tag": processPlanLogInput.Tag}
		err = c.Find(query).Sort("-updatetime").All(&logs)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get process plan log info failed, process plan id: %s, err: %v", processPlanLogInput.ProcessName, err)
			return
		}
		totalNum, err = c.Find(query).Count()
		if err != nil {
			err = status.Errorf(codes.Aborted, "get process plan log total num failed, err : %v", err)
			return
		}
		if logs != nil {
			var planInfoOutput *siber.PlanLog
			for i, m := range logs {
				for j, n := range m.PlanLog {
					planInfoOutput, err = dao.SelectPlanLogDetail(ctx, &siber.PlanLog{PlanLogId: n.PlanLogId})
					logs[i].PlanLog[j] = planInfoOutput
				}
			}
		}
	}

	return &logs, totalNum, err
}

//日志插入
func (dao *Dao) InsertProcessPlanLog(ctx context.Context, processPlanLogInput *siber.ProcessPlanLogInfo) (err error) {
	err = isInputNull(ctx, processPlanLogInput)
	processPlanLogInput.UpdateTime = time.Now().Unix()
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionProcessPlanLog)
	if err != nil {
		return
	}
	_, err = c.Upsert(bson.M{"processname": processPlanLogInput.ProcessName, "tag": processPlanLogInput.Tag}, processPlanLogInput)
	if err != nil {
		err = status.Errorf(codes.Aborted, "update process plan log failed, err: %v", err)
		return
	}
	return
}
