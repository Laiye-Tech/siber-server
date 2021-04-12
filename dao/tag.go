/**
* @Author: TongTongLiu
* @Date: 2019/10/12 6:38 下午
**/

package dao

import (
	"context"
	bson2 "github.com/globalsign/mgo/bson"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

type tagInfo struct {
	Id      string        `bson:"_id,omitempty"`
	TagInfo siber.TagInfo `bson:",inline"`
}

func contains(values []string, ExceptValue string) bool {
	for _, value := range values {
		if value == ExceptValue {
			return true
		}
	}
	return false
}
func (dao *Dao) InsertTag(ctx context.Context, tagInfoInput *siber.TagInfo) (tagInfoOutput *siber.TagInfo, err error) {
	err = isInputNull(ctx, tagInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionTag)
	if err != nil {
		return
	}
	tagInfoInput.InsertTime = time.Now().Unix()
	tagInfoInput.UpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	tagInfoInput.TagId = mongoID
	tagInfo := &tagInfo{
		Id:      mongoID,
		TagInfo: *tagInfoInput,
	}
	err = c.Insert(tagInfo)
	if err != nil {
		if mgo.IsDup(err) {
			xzap.Logger(ctx).Info("failed to persistence tag")
			err = status.Errorf(codes.AlreadyExists, "tag name is a duplicate")
			return
		}
		xzap.Logger(ctx).Error("failed to persistence tag", zap.Any("err", err))
		err = status.Errorf(codes.InvalidArgument, "failed to insert Tag")
		return
	}
	tagInfoOutput, err = dao.SelectTag(ctx, tagInfoInput)
	return
}

func (dao *Dao) SelectTag(ctx context.Context, tagInfoInput *siber.TagInfo) (tagInfoOutput *siber.TagInfo, err error) {
	err = isInputNull(ctx, tagInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionTag)
	if err != nil {
		return
	}
	tagInfoOutput = new(siber.TagInfo)
	if tagInfoInput.TagId != "" {
		// 当存在ID时，根据ID进行查询
		err = c.Find(bson.M{"tagid": tagInfoInput.TagId}).One(&tagInfoOutput)
		if err != nil {
			xzap.Logger(ctx).Error("get tag info failed, tag", zap.Any("err:", err))
			return
		}
	} else {
		// 当ID不存在时，根据TagName 进行查询
		err = c.Find(bson.M{"tagname": tagInfoInput.TagName}).One(&tagInfoOutput)
		if err != nil {
			return
		}
	}
	return
}

func (dao *Dao) UpdateTag(ctx context.Context, tagInfoInput *siber.TagInfo) (tagInfoOutput *siber.TagInfo, err error) {
	err = isInputNull(ctx, tagInfoInput)
	if err != nil {
		return
	}
	if tagInfoInput.TagId == "" {
		err = status.Errorf(codes.InvalidArgument, "TagID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionTag)
	if err != nil {
		return
	}
	tagInfoInput.UpdateTime = time.Now().Unix()
	err = c.Update(bson.M{"tagid": tagInfoInput.TagId}, bson.M{"$set": tagInfoInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "tag name is a duplicate")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to update Tag")
		return
	}
	tagInfoOutput, err = dao.SelectTag(ctx, tagInfoInput)
	return
}

func (dao *Dao) DeleteTag(ctx context.Context, tagInfoInput *siber.TagInfo) (tagInfoOutput *siber.TagInfo, err error) {
	err = isInputNull(ctx, tagInfoInput)
	if err != nil {
		return
	}
	if tagInfoInput.TagId == "" {
		err = status.Errorf(codes.InvalidArgument, "TagID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionTag)
	if err != nil {
		return
	}
	selector := bson.M{"tagid": tagInfoInput.TagId}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	tagInfoOutput = new(siber.TagInfo)
	return
}

func (dao *Dao) ListTag(ctx context.Context, filterInfo *siber.FilterInfo) (TagList *[]*siber.TagInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionTag)
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
	if _, ok := filterInfo.FilterContent["content"]; ok {
		var filter = make(map[string]string)
		filter["tagname"] = filterInfo.FilterContent["content"]
		var orQueries []bson.M
		for k, v := range filter {
			orQueries = append(orQueries, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		andQuery = append(andQuery, bson.M{"$or": orQueries})
	}
	andQuery = append(andQuery, bson.M{"invaliddate": 0})
	query = bson.M{"$and": andQuery}
	var Tags []*siber.TagInfo
	err = c.Find(query).Sort("-updatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&Tags)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get Tag list failed, err : %v", err)
		return
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get Tag total num failed, err : %v", err)
		return
	}
	return &Tags, totalNum, err
}
func (dao *Dao) SelectTags(ctx context.Context, tagIds []string) (tagNames []string, err error) {
	if len(tagIds) <= 0 {
		return
	}
	var _tagNames []string
	tagList := new([]*siber.TagInfo)
	c, err := getMongoCollection(ctx, CollectionTag)
	if err != nil {
		return nil, err
	}
	err = c.Find(bson.M{"tagid": bson.M{"$in": tagIds}}).All(tagList)
	if err != nil {
		return nil, err
	}
	for _, v := range *tagList {
		_tagNames = append(_tagNames, v.TagName)
	}
	return _tagNames, nil
}

func (dao *Dao) InsertTags(ctx context.Context, tagNames []string) (tagIds []string, err error) {
	if len(tagNames) <= 0 {
		return
	}
	var _tagNames []string
	var _tagIds []string
	var tempTagInfos []interface{}
	tagInfoList := new([]*siber.TagInfo)
	c, err := getMongoCollection(ctx, CollectionTag)
	if err != nil {
		return nil, err
	}
	err = c.Find(bson.M{"tagname": bson.M{"$in": tagNames}}).All(tagInfoList)
	if err != nil {
		return nil, err
	}
	for _, v := range *tagInfoList {
		_tagNames = append(_tagNames, v.TagName)
	}
	for _, v := range tagNames {
		if contains(_tagNames, v) {
			continue
		}
		mongoID := bson2.NewObjectId().Hex()
		tempTagInfos = append(tempTagInfos, &tagInfo{
			Id: mongoID,
			TagInfo: siber.TagInfo{
				TagId:      mongoID,
				TagName:    v,
				InsertTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
		})
	}
	if len(tempTagInfos) > 0 {
		err = c.Insert(tempTagInfos...)
		if err != nil {
			if mgo.IsDup(err) {
				xzap.Logger(ctx).Info("failed to persistence tag")
				err = status.Errorf(codes.AlreadyExists, "tag name is a duplicate")
				return
			}
			return nil, err
		}
	}
	err = c.Find(bson.M{"tagname": bson.M{"$in": tagNames}}).All(tagInfoList)
	if err != nil {
		return nil, err
	}
	for _, vv := range *tagInfoList {
		_tagIds = append(_tagIds, vv.TagId)
	}
	return _tagIds, nil
}
