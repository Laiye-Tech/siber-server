/**
* @Author: TongTongLiu
* @Date: 2019/10/12 6:38 下午
**/

package dao

import (
	"context"
	bson2 "github.com/globalsign/mgo/bson"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

type envInfo struct {
	Id      string        `bson:"_id,omitempty"`
	EnvInfo siber.EnvInfo `bson:",inline"`
}

func (dao *Dao) InsertEnv(ctx context.Context, envInfoInput *siber.EnvInfo) (envInfoOutput *siber.EnvInfo, err error) {
	err = isInputNull(ctx, envInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionEnv)
	if err != nil {
		return
	}
	envInfoInput.InsertTime = time.Now().Unix()
	envInfoInput.UpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	envInfoInput.EnvId = mongoID
	envInfo := &envInfo{
		Id:      mongoID,
		EnvInfo: *envInfoInput,
	}
	err = c.Insert(envInfo)
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "env name is a duplicate")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "insert Env failed, err: %v", err)
		return
	}
	envInfoOutput, err = dao.SelectEnv(ctx, envInfoInput)
	return
}

func (dao *Dao) SelectEnv(ctx context.Context, envInfoInput *siber.EnvInfo) (envInfoOutput *siber.EnvInfo, err error) {
	err = isInputNull(ctx, envInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionEnv)
	if err != nil {
		return
	}
	envInfoOutput = new(siber.EnvInfo)
	if envInfoInput.EnvId != "" {
		// 当存在ID时，根据ID进行查询
		err = c.Find(bson.M{"envid": envInfoInput.EnvId}).One(&envInfoOutput)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "get env info failed, env:%v, err:%v", envInfoInput, err)
			return
		}
	} else {
		// 当ID不存在时，根据EnvName 进行查询
		err = c.Find(bson.M{"envname": envInfoInput.EnvName, "invaliddate": 0}).One(&envInfoOutput)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "get env info failed, env:%v, err:%v", envInfoInput, err)
			return
		}
	}
	return
}

func (dao *Dao) UpdateEnv(ctx context.Context, envInfoInput *siber.EnvInfo) (envInfoOutput *siber.EnvInfo, err error) {
	err = isInputNull(ctx, envInfoInput)
	if err != nil {
		return
	}
	if envInfoInput.EnvId == "" {
		err = status.Errorf(codes.InvalidArgument, "EnvID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionEnv)
	if err != nil {
		return
	}
	envInfoInput.UpdateTime = time.Now().Unix()
	err = c.Update(bson.M{"envid": envInfoInput.EnvId}, bson.M{"$set": envInfoInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "env name is a duplicate")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "update Env failed, err: %v", err)
		return
	}
	envInfoOutput, err = dao.SelectEnv(ctx, envInfoInput)
	return
}

func (dao *Dao) DeleteEnv(ctx context.Context, envInfoInput *siber.EnvInfo) (envInfoOutput *siber.EnvInfo, err error) {
	err = isInputNull(ctx, envInfoInput)
	if err != nil {
		return
	}
	if envInfoInput.EnvId == "" {
		err = status.Errorf(codes.InvalidArgument, "EnvID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionEnv)
	if err != nil {
		return
	}
	selector := bson.M{"envid": envInfoInput.EnvId}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	envInfoOutput = new(siber.EnvInfo)

	return
}

func (dao *Dao) ListEnv(ctx context.Context, filterInfo *siber.FilterInfo) (EnvList *[]*siber.EnvInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionEnv)
	if err != nil {
		return
	}
	var query bson.M
	var andQuery []bson.M

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
	if _, ok := filterInfo.FilterContent["method"]; ok {
		if filterInfo.FilterContent["method"] == "instance" {
			//andQuery = append(andQuery, bson.M{"instance": bson.M{"$ne": nil}})
			andQuery = append(andQuery, bson.M{"envmode": "instance"})

		}
		if filterInfo.FilterContent["method"] == "interface" {
			andQuery = append(andQuery, bson.M{"envmode": "interface"})
		}
	}
	if _, ok := filterInfo.FilterContent["content"]; ok {
		var filter = make(map[string]string)
		filter["envname"] = filterInfo.FilterContent["content"]
		var orQueries []bson.M
		for k, v := range filter {
			orQueries = append(orQueries, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		andQuery = append(andQuery, bson.M{"$or": orQueries})
	}
	andQuery = append(andQuery, bson.M{"invaliddate": 0})
	query = bson.M{"$and": andQuery}
	var Envs []*siber.EnvInfo
	err = c.Find(query).Sort("-updatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&Envs)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get Env list failed, err : %v", err)
		return
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get Env total num failed, err : %v", err)
		return
	}
	return &Envs, totalNum, err
}
