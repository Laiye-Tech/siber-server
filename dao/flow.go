/**
* @Author: TongTongLiu
* @Date: 2019-09-17 10:55
**/

package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

type flowInfo struct {
	Id       string         `bson:"_id,omitempty"`
	FlowInfo siber.FlowInfo `bson:",inline"`
}

func (dao *Dao) SelectFlow(ctx context.Context, flowInfoInput *siber.FlowInfo) (flowInfoOutput *siber.FlowInfo, err error) {
	err = isInputNull(ctx, flowInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionFlow)
	if err != nil {
		return
	}
	flowInfoOutput = new(siber.FlowInfo)
	if flowInfoInput.FlowId != "" {
		// 当存在ID时，根据ID进行查询
		err = c.Find(bson.M{"flowid": flowInfoInput.FlowId}).One(&flowInfoOutput)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get flow info failed, flowID:%s, err:%v", flowInfoInput.FlowId, err)
			return
		}
	} else {
		// 当ID不存在时，根据FlowName 进行查询
		err = c.Find(bson.M{"flowname": flowInfoInput.FlowName, "invaliddate": 0}).One(&flowInfoOutput)
		if err != nil {
			err = status.Errorf(codes.Aborted, "get flow info failed, flowID:%s, err:%v", flowInfoInput.FlowName, err)
			return
		}
	}
	return
}

func (dao *Dao) InsertFlow(ctx context.Context, flowInfoInput *siber.FlowInfo) (flowInfoOutput *siber.FlowInfo, err error) {
	err = isInputNull(ctx, flowInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionFlow)
	if err != nil {
		return
	}
	flowInfoInput.InsertTime = time.Now().Unix()
	flowInfoInput.UpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	flowInfoInput.FlowId = mongoID
	flowInfo := &flowInfo{
		Id:       mongoID,
		FlowInfo: *flowInfoInput,
	}
	err = c.Insert(flowInfo)
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "flow name is a duplicate")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to insert Flow, err:%v", err)
		return
	}
	flowInfoOutput, err = dao.SelectFlow(ctx, flowInfoInput)
	return
}

func (dao *Dao) UpdateFlow(ctx context.Context, flowInfoInput *siber.FlowInfo) (flowInfoOutput *siber.FlowInfo, err error) {
	err = isInputNull(ctx, flowInfoInput)
	if err != nil {
		return
	}
	if flowInfoInput.FlowId == "" {
		err = status.Errorf(codes.InvalidArgument, "FlowID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionFlow)
	if err != nil {
		return
	}
	flowInfoInput.UpdateTime = time.Now().Unix()
	err = c.Update(bson.M{"flowid": flowInfoInput.FlowId}, bson.M{"$set": flowInfoInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "flow name is a duplicate")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to update Flow, err :%v", err)
		return
	}
	flowInfoOutput, err = dao.SelectFlow(ctx, flowInfoInput)
	return
}

func (dao *Dao) DeleteFlow(ctx context.Context, flowInfoInput *siber.FlowInfo) (flowInfoOutput *siber.FlowInfo, err error) {
	err = isInputNull(ctx, flowInfoInput)
	if err != nil {
		return
	}
	if flowInfoInput.FlowId == "" {
		err = status.Errorf(codes.InvalidArgument, "FlowID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionFlow)
	if err != nil {
		return
	}
	selector := bson.M{"flowid": flowInfoInput.FlowId}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	flowInfoOutput = new(siber.FlowInfo)

	return
}

func (dao *Dao) ListFlow(ctx context.Context, filterInfo *siber.FilterInfo) (flowList *[]*siber.FlowInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionFlow)
	if err != nil {
		return
	}
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

	// 模糊匹配
	if _, ok := filterInfo.FilterContent["content"]; ok {
		var filter = make(map[string]string)
		filter["flowname"] = filterInfo.FilterContent["content"]
		var orQuerys []bson.M
		for k, v := range filter {
			orQuerys = append(orQuerys, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		var andQuery []bson.M
		andQuery = append(andQuery, bson.M{"invaliddate": 0}, bson.M{"automatic": bson.M{"$ne": 1}})
		andQuery = append(andQuery, bson.M{"$or": orQuerys})
		query = bson.M{"$and": andQuery}
	}

	// 精准匹配
	if _, ok := filterInfo.FilterContent[CaseNode]; ok {
		andQuery := []bson.M{
			bson.M{"caselist": bson.M{"$in": []string{filterInfo.FilterContent[CaseNode]}}},
			bson.M{"invaliddate": 0},
			bson.M{"automatic": bson.M{"$ne": 1}},
		}
		query = bson.M{"$and": andQuery}
	}
	if query == nil {
		var andQuerys []bson.M
		andQuerys = append(andQuerys, bson.M{"invaliddate": 0}, bson.M{"automatic": bson.M{"$ne": 1}})
		query = bson.M{"$and": andQuerys}
	}
	var flows []*siber.FlowInfo
	err = c.Find(query).Sort("-updatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&flows)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get flow list failed, err : %v", err)
		return
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan total num failed, err : %v", err)
		return
	}
	return &flows, totalNum, err
}
