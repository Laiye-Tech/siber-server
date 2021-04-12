/**
* @Author: TongTongLiu
* @Date: 2019-09-17 10:56
**/

package dao

import (
	"context"
	"strconv"
	"time"

	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

type planInfo struct {
	Id       string         `bson:"_id,omitempty"`
	PlanInfo siber.PlanInfo `bson:",inline"`
}

func (dao *Dao) SelectPlan(ctx context.Context, planInfoInput *siber.PlanInfo) (planInfoOutput *siber.PlanInfo, err error) {
	err = isInputNull(ctx, planInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionPlan)
	if err != nil {
		return
	}
	planInfoOutput = new(siber.PlanInfo)
	if planInfoInput.PlanId != "" {
		// 当存在ID时，根据ID进行查询
		err = c.Find(bson.M{"planid": planInfoInput.PlanId}).One(&planInfoOutput)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get plan info failed, planID: %s, err: %v", planInfoInput.PlanId, err)
			return
		}
	} else {
		// 当ID不存在时，根据PlanName 进行查询
		err = c.Find(bson.M{"planname": planInfoInput.PlanName, "invaliddate": 0}).One(&planInfoOutput)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get plan info failed, planName: %s, err: %v", planInfoInput.PlanName, err)
			return
		}
	}

	return
}

func (dao *Dao) InsertPlan(ctx context.Context, planInfoInput *siber.PlanInfo) (planInfoOutput *siber.PlanInfo, err error) {
	err = isInputNull(ctx, planInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionPlan)
	if err != nil {
		return
	}
	planInfoInput.InsertTime = time.Now().Unix()
	planInfoInput.UpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	planInfoInput.PlanId = mongoID
	planInfo := &planInfo{
		Id:       mongoID,
		PlanInfo: *planInfoInput,
	}
	err = c.Insert(planInfo)
	if err != nil {
		err = status.Errorf(codes.Aborted, "persistence plan failed, err: %v", err)
		return
	}

	planInfoOutput, err = dao.SelectPlan(ctx, planInfoInput)
	return
}

func (dao *Dao) UpdatePlan(ctx context.Context, planInfoInput *siber.PlanInfo) (planInfoOutput *siber.PlanInfo, err error) {
	err = isInputNull(ctx, planInfoInput)
	if err != nil {
		return
	}
	if planInfoInput.PlanId == "" {
		err = status.Errorf(codes.InvalidArgument, "PlanID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionPlan)
	if err != nil {
		return
	}
	planInfoInput.UpdateTime = time.Now().Unix()
	err = c.Update(bson.M{"planid": planInfoInput.PlanId}, bson.M{"$set": planInfoInput})
	if err != nil {
		err = status.Errorf(codes.Aborted, "update plan failed, err: %v", err)
		return
	}
	planInfoOutput, err = dao.SelectPlan(ctx, planInfoInput)
	return
}

func (dao *Dao) DeletePlan(ctx context.Context, planInfoInput *siber.PlanInfo) (planInfoOutput *siber.PlanInfo, err error) {
	err = isInputNull(ctx, planInfoInput)
	if err != nil {
		return
	}
	if planInfoInput.PlanId == "" {
		err = status.Errorf(codes.InvalidArgument, "PlanID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionPlan)
	if err != nil {
		return
	}
	selector := bson.M{"_id": planInfoInput.PlanId}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	planInfoOutput = new(siber.PlanInfo)
	return
}

func (dao *Dao) ListPlan(ctx context.Context, filterInfo *siber.FilterInfo) (planList *[]*siber.PlanInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionPlan)
	if err != nil {
		return
	}
	var plans []*siber.PlanInfo
	var query bson.M
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
	var andQueries []bson.M
	// 项目启动时添加 cron 的搜索
	if _, ok := filterInfo.FilterContent["action"]; ok {
		if "cron" == filterInfo.FilterContent["action"] {
			andQueries = append(andQueries, bson.M{"triggercondition": bson.M{"$elemMatch": bson.M{"triggercron": bson.M{"$ne": ""}}}})
			andQueries = append(andQueries, bson.M{"triggercondition": bson.M{"$elemMatch": bson.M{"triggercron": bson.M{"$ne": nil}}}})
			andQueries = append(andQueries, bson.M{"invaliddate": 0}, bson.M{"automatic": bson.M{"$ne": 1}})
		}

	}

	// 模糊搜索
	if _, ok := filterInfo.FilterContent["content"]; ok {
		var filter = make(map[string]string)
		filter["planname"] = filterInfo.FilterContent["content"]
		filter["environmentname"] = filterInfo.FilterContent["content"]
		var orQueries []bson.M
		for k, v := range filter {
			orQueries = append(orQueries, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		andQueries = append(andQueries, bson.M{"invaliddate": 0}, bson.M{"automatic": bson.M{"$ne": 1}})
		andQueries = append(andQueries, bson.M{"$or": orQueries})
		query = bson.M{"$and": andQueries}
		err = c.Find(query).Sort("-updatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&plans)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get plan list failed, err : %v", err)
			return
		}
		totalNum, err = c.Find(query).Count()
		return &plans, totalNum, err
	}

	// 精准匹配:根据flow查询plan
	if _, ok := filterInfo.FilterContent[FlowNode]; ok {
		andQueries := []bson.M{
			bson.M{"flowlist": bson.M{"$in": []string{filterInfo.FilterContent[FlowNode]}}},
			bson.M{"invaliddate": 0},
			bson.M{"automatic": bson.M{"$ne": 1}},
		}
		query = bson.M{"$and": andQueries}
	}

	// 精准匹配：根据service_name 查询plan
	if serviceName, ok := filterInfo.FilterContent["services"]; ok {
		andQueries := []bson.M{
			bson.M{"services": serviceName},
			bson.M{"invaliddate": 0},
			bson.M{"automatic": bson.M{"$ne": 1}},
		}
		query = bson.M{"$and": andQueries}
	}
	if bindServiceName, ok := filterInfo.FilterContent["bind_services"]; ok {
		if bindServerEnv, ok := filterInfo.FilterContent["environment_name"]; ok {
			andQueries := []bson.M{
				bson.M{"triggercondition": bson.M{"$elemMatch": bson.M{"triggerservicelist": bindServiceName, "environmentname": bindServerEnv}}},
				bson.M{"invaliddate": 0},
				bson.M{"automatic": bson.M{"$ne": 1}},
			}
			query = bson.M{"$and": andQueries}

		}
	}
	if query == nil {
		var andQueries []bson.M
		andQueries = append(andQueries, bson.M{"invaliddate": 0}, bson.M{"automatic": bson.M{"$ne": 1}})
		query = bson.M{"$and": andQueries}
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan total num failed, err : %v", err)
		return
	}
	err = c.Find(query).All(&plans)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan list failed, err : %v", err)
		return
	}
	return &plans, totalNum, err
}

func StatPlan(ctx context.Context, query bson2.M) (int, error) {
	c, err := getMongoCollection(ctx, CollectionLogPlan)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get mongo collection error: %v", err)
		return 0, err
	}
	num, err := c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "find total num of plan from the collection error: %v", err)
		return 0, err
	}
	return num, nil
}

func GenerateSuccessfulPlanNumQuery(date uint64, endTime uint64) bson2.D {
	query := bson2.D{
		{"planstatus", 2},
		{"dbinserttime", bson2.D{
			{"$lt", endTime - 86400*(date-1)},
			{"$gt", endTime - 86400*(date)},
		}},
	}
	return query
}

func GenerateTotalPlanNumQuery(date uint64, endTime uint64) bson2.D {
	query := bson2.D{
		{"dbinserttime", bson2.D{
			{"$lt", endTime - 86400*(date-1)},
			{"$gt", endTime - 86400*(date)},
		}},
	}
	return query
}
