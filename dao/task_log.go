/**
* @Author: TongTongLiu
* @Date: 2019/10/16 2:43 下午
**/

package dao

import (
	"context"
	"encoding/json"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"time"

	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

type CaseLog struct {
	Id      string        `bson:"_id,omitempty"`
	CaseLog siber.CaseLog `bson:",inline"`
}

type FlowLog struct {
	Id      string        `bson:"_id,omitempty"`
	FlowLog siber.FlowLog `bson:",inline"`
}

type PlanLog struct {
	Id      string        `bson:"_id,omitempty"`
	PlanLog siber.PlanLog `bson:",inline"`
}

// 有则更新，无则插入
func (dao *Dao) UpsertCaseLog(ctx context.Context, log *siber.CaseLog) (resp *siber.CaseLog, err error) {
	c, err := getMongoCollection(ctx, CollectionLogCase)
	if err != nil {
		return
	}
	caseLog := &CaseLog{}
	if log.CaseLogId == "" {
		mongoID := bson2.NewObjectId().Hex()
		log.CaseLogId = mongoID
		caseLog.Id = mongoID
		resp = new(siber.CaseLog)
		resp.CaseLogId = caseLog.Id
	}
	//  TODO:临时debug,待清理
	bLog, err := json.Marshal(*log)
	if err != nil {
		return
	}
	xzap.Logger(ctx).Info("blog:" + string(bLog))
	caseLog.CaseLog = *log
	_, err = c.Upsert(bson.M{"caselogid": log.CaseLogId}, bson.M{"$set": caseLog})
	if err != nil {
		err = status.Errorf(codes.Aborted, "failed to persistence case log:%v, err:%v ", log, err)
		return
	}

	return
}

func (dao *Dao) InsertFlowLog(ctx context.Context, log *siber.FlowLog) (resp *siber.FlowLog, err error) {
	err = isInputNull(ctx, log)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "Insert flow log get nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionLogFlow)
	if err != nil {
		return
	}
	log.DbUpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	if log.FlowLogId == "" {
		log.FlowLogId = mongoID
		flowLog := &FlowLog{
			Id:      mongoID,
			FlowLog: *log,
		}
		err = c.Insert(flowLog)
	} else {
		err = c.Update(bson.M{"flowlogid": log.FlowLogId}, log)
	}
	if err != nil {
		err = status.Errorf(codes.Aborted, "failed to persistence flow log. log:%v, err:%v", log, err)
		return
	}

	return
}

func (dao *Dao) InsertPlanLog(ctx context.Context, log *siber.PlanLog) (resp *siber.PlanLog, err error) {
	err = isInputNull(ctx, log)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "Insert Plan log get nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionLogPlan)
	if err != nil {
		return
	}
	log.DbUpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	if log.PlanLogId == "" {
		log.PlanLogId = mongoID
		planLog := &PlanLog{
			Id:      mongoID,
			PlanLog: *log,
		}
		err = c.Insert(planLog)
	} else {
		err = c.Update(bson.M{"planlogid": log.PlanLogId}, log)
	}

	if err != nil {
		err = status.Errorf(codes.Aborted, "failed to persistence plan log, err:%v, log info:%v ", err, log)
		return
	}
	resp = new(siber.PlanLog)
	resp.PlanLogId = mongoID
	return resp, err
}

func (dao *Dao) SelectCaseLogDetail(ctx context.Context, req *siber.CaseLog) (log *siber.CaseLog, err error) {
	if req == nil || req.CaseLogId == "" {
		err = status.Errorf(codes.InvalidArgument, "select case log detail failed, req or CaseLogId is nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionLogCase)
	if err != nil {
		return
	}
	// 仅支持根据caseLogID查询
	log = new(siber.CaseLog)
	err = c.Find(bson.M{"caselogid": req.CaseLogId}).One(&log)
	if err != nil {
		err = status.Errorf(codes.Aborted, "select case log detail failed, err:%v", err)
	}
	return
}

func (dao *Dao) SelectPlanLogDetail(ctx context.Context, req *siber.PlanLog) (planLog *siber.PlanLog, err error) {
	if req == nil || req.PlanLogId == "" {
		err = status.Errorf(codes.InvalidArgument, "select case log detail failed, req or CaseLogId is nil")
		return
	}
	c, err := getMongoCollection(ctx, CollectionLogPlan)
	if err != nil {
		return
	}
	// 仅支持根据planLogID查询
	err = c.Find(bson.M{"planlogid": req.PlanLogId}).One(&planLog)
	if err != nil {
		err = status.Errorf(codes.Aborted, "select plan log detail failed, err:%v", err)
	}
	return planLog, err
}

func (dao *Dao) ListCaseLog(ctx context.Context, req *siber.CaseLog) (logList *[]*siber.CaseLog, err error) {
	c, err := getMongoCollection(ctx, CollectionLogCase)
	if err != nil {
		return
	}
	var caseLogList []*siber.CaseLog
	// 拼接查询条件
	cond := make(map[string]interface{})
	if req.PlanLogId != "" {
		cond["planlogid"] = req.PlanLogId
	}
	if req.FlowLogId != "" {
		cond["flowlogid"] = req.FlowLogId
	}
	if req.PlanId != "" {
		cond["planid"] = req.PlanId
	}
	if req.FlowId != "" {
		cond["flowid"] = req.FlowId
	}
	if req.CaseId != "" {
		cond["caseid"] = req.CaseId
	}
	err = c.Find(cond).Select(bson.M{"caselogid": 1, "planname": 1, "flowname": 1, "casename": 1, "dbinserttime": -1, "dbupdatetime": -1, "casestatus": 1, "methodname": 1, "errcontent": 1}).All(&caseLogList)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get case list failed, err : %v", err)
	}
	return &caseLogList, err
}

func (dao *Dao) ListFlowLog(ctx context.Context, req *siber.FlowLog) (logList *[]*siber.FlowLog, err error) {
	c, err := getMongoCollection(ctx, CollectionLogFlow)
	if err != nil {
		return
	}
	var flowLogList []*siber.FlowLog
	// 拼接查询条件
	cond := make(map[string]interface{})
	if req.PlanLogId != "" {
		cond["planlogid"] = req.PlanLogId
	}
	err = c.Find(cond).All(&flowLogList)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get flow list failed, err : %v", err)
	}
	return &flowLogList, err
}

func (dao *Dao) ListPlanLog(ctx context.Context, req *siber.ListPlanLogRequest) (logList *[]*siber.PlanLog, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionLogPlan)
	if err != nil {
		return
	}
	var query bson.M
	var andQuerys []bson.M
	var planLogList []*siber.PlanLog
	page := defaultPage
	pageSize := defaultPageSize
	if req.Page != 0 {
		page = int(req.Page)
	}
	if req.PageSize != 0 {
		pageSize = int(req.PageSize)
	}
	if req.PlanId != "" {
		andQuerys = append(andQuerys, bson.M{"planid": req.PlanId})
	}
	if req.Params != nil {
		if req.Params.PlanName != "" {
			andQuerys = append(andQuerys, bson.M{"planname": bson.M{"$regex": req.Params.PlanName, "$options": "$i"}})
		}
		if req.Params.InterfaceType != nil && len(req.Params.InterfaceType) != 0 {
			var orQuerys []bson.M
			for _, v := range req.Params.InterfaceType {
				orQuerys = append(orQuerys, bson.M{"interfacetype": v})
			}
			andQuerys = append(andQuerys, bson.M{"$or": orQuerys})
		}
		if req.Params.Trigger != nil && len(req.Params.Trigger) != 0 {
			var orQuerys []bson.M
			for _, v := range req.Params.Trigger {
				orQuerys = append(orQuerys, bson.M{"trigger": v})
			}
			andQuerys = append(andQuerys, bson.M{"$or": orQuerys})
		}
		if req.Params.PlanStatus != nil && len(req.Params.PlanStatus) != 0 {
			var orQuerys []bson.M
			for _, v := range req.Params.PlanStatus {
				orQuerys = append(orQuerys, bson.M{"planstatus": v})
			}
			andQuerys = append(andQuerys, bson.M{"$or": orQuerys})
		}
		if req.Params.VersionControl != nil && len(req.Params.VersionControl) != 0 {
			var orQuerys []bson.M
			for _, v := range req.Params.VersionControl {
				orQuerys = append(orQuerys, bson.M{"versioncontrol": v})
			}
			andQuerys = append(andQuerys, bson.M{"$or": orQuerys})
		}
		if req.Params.EnvironmentName != nil && len(req.Params.EnvironmentName) != 0 {
			var orQuerys []bson.M
			for _, v := range req.Params.EnvironmentName {
				orQuerys = append(orQuerys, bson.M{"environmentname": v})
			}
			andQuerys = append(andQuerys, bson.M{"$or": orQuerys})
		}
	}
	if len(andQuerys) != 0 {
		query = bson.M{"$and": andQuerys}
	}
	err = c.Find(query).Sort("-dbupdatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&planLogList)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan list failed, err : %v", err)
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan log total num failed, err : %v", err)
		return
	}
	return &planLogList, totalNum, err
}
