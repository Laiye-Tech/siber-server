/**
* @Author: TongTongLiu
* @Date: 2019-09-17 10:52
**/

package dao

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

// TODO: 批量的增删改查
// TODO: 获取case 的依赖关系

type CheckerStandard struct {
	Key      string
	Relation string
	Content  interface{}
}

type caseInfo struct {
	Id       string         `bson:"_id,omitempty"`
	CaseInfo siber.CaseInfo `bson:",inline"`
}

func (dao *Dao) SelectCase(ctx context.Context, caseInfoInput *siber.CaseInfo) (caseInfoOutput *siber.CaseInfo, err error) {
	err = isInputNull(ctx, caseInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	caseInfoOutput = new(siber.CaseInfo)
	if caseInfoInput.CaseId != "" {
		// 当存在ID时，根据ID进行查询
		err = c.Find(bson.M{"caseid": caseInfoInput.CaseId}).One(&caseInfoOutput)
		if err != nil {
			return nil, err
		}
	} else {
		// 当ID不存在时，根据CaseName 进行查询
		err = c.Find(bson.M{"casename": caseInfoInput.CaseName, "invaliddate": 0}).One(&caseInfoOutput)
		if err != nil {
			return nil, err
		}
	}
	err = c.Find(bson.M{"caseid": caseInfoInput.CaseId}).One(&caseInfoOutput)
	if err != nil {
		return nil, err
	}
	TagNames, err := dao.SelectTags(ctx, caseInfoOutput.CaseTags)
	if err != nil {
		return nil, err
	}
	if len(TagNames) > 0 {
		caseInfoOutput.CaseTags = TagNames
	}
	return caseInfoOutput, err
}

func (dao *Dao) SelectCaseByIDList(ctx context.Context, idList *[]string) (caseList *[]*siber.ManageFlowCaseSub, err error) {
	if idList == nil || len(*idList) == 0 {
		xzap.Logger(ctx).Info("SelectCaseByIDList get nil idList")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	var newCaseList []*siber.ManageFlowCaseSub
	var tempCaseList []*siber.ManageFlowCaseSub

	err = c.Find(bson.M{"caseid": bson.M{"$in": idList}}).All(&tempCaseList)
	for _, v := range *idList {
		for _, t := range tempCaseList {
			if t.CaseId == v {
				caseDetail := new(siber.ManageFlowCaseSub)
				caseDetail.CaseId = v
				caseDetail.CaseName = t.CaseName
				newCaseList = append(newCaseList, caseDetail)
			}
		}
	}
	caseList = &newCaseList
	if err != nil {
		xzap.Logger(ctx).Error("SelectCaseByIDList failed", zap.Any("err", err))
	}
	return
}

func (dao *Dao) InsertCase(ctx context.Context, caseInfoInput *siber.CaseInfo) (caseInfoOutput *siber.CaseInfo, err error) {
	err = isInputNull(ctx, caseInfoInput)
	if err != nil {
		return
	}
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	caseInfoInput.InsertTime = time.Now().Unix()
	caseInfoInput.UpdateTime = time.Now().Unix()
	mongoID := bson2.NewObjectId().Hex()
	caseInfoInput.CaseId = mongoID
	tags, err := dao.InsertTags(ctx, caseInfoInput.CaseTags)
	if err != nil {
		return nil, err
	}
	caseInfoInput.CaseTags = tags
	caseInfo := &caseInfo{
		Id:       mongoID,
		CaseInfo: *caseInfoInput,
	}
	err = c.Insert(caseInfo)
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "case name is a duplicate")
			return
		}
		xzap.Logger(ctx).Error("failed to insert case", zap.Any("caseInfoInput", caseInfoInput), zap.Any("err", err))
		err = status.Errorf(codes.InvalidArgument, "failed to insert case")
		return
	}
	caseInfoOutput, err = NewDao().SelectCase(ctx, caseInfoInput)
	return
}

// 和mysql不同，Update函数只能用来修改单条记录，即使条件能匹配多条记录，也只会修改第一条匹配的记录。
// 所以必须通过主键进行update
func (dao *Dao) UpdateCase(ctx context.Context, caseInfoInput *siber.CaseInfo) (caseInfoOutput *siber.CaseInfo, err error) {
	err = isInputNull(ctx, caseInfoInput)
	if err != nil {
		return
	}
	if caseInfoInput.CaseId == "" {
		xzap.Logger(ctx).Error("try to update case but CaseId is null")
		err = status.Errorf(codes.InvalidArgument, "caseID is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	caseInfoInput.UpdateTime = time.Now().Unix()
	tags, err := dao.InsertTags(ctx, caseInfoInput.CaseTags)
	if err != nil {
		return nil, err
	}
	caseInfoInput.CaseTags = tags
	err = c.Update(bson.M{"caseid": caseInfoInput.CaseId}, bson.M{"$set": caseInfoInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "case name is a duplicate")
			return
		}
		xzap.Logger(ctx).Warn("failed to update case: ", zap.Any("caseInfoInput", caseInfoInput), zap.Any("err", err))
		err = status.Errorf(codes.InvalidArgument, "failed to update case")
		return
	}
	caseInfoOutput, err = dao.SelectCase(ctx, caseInfoInput)
	return
}

func (dao *Dao) DeleteCase(ctx context.Context, caseInfoInput *siber.CaseInfo) (caseInfoOutput *siber.CaseInfo, err error) {
	err = isInputNull(ctx, caseInfoInput)
	if err != nil {
		xzap.Logger(ctx).Warn("DeleteCase got nil input, please check")
		return
	}
	if caseInfoInput.CaseId == "" {
		xzap.Logger(ctx).Warn("try to delete case but CaseId is null")
		err = status.Errorf(codes.InvalidArgument, "CaseId is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	selector := bson.M{"caseid": caseInfoInput.CaseId}
	data := bson.M{"$set": bson.M{"invaliddate": time.Now().Unix()}}
	err = c.Update(selector, data)
	caseInfoOutput = new(siber.CaseInfo)

	return
}

func (dao *Dao) ListCase(ctx context.Context, filterInfo *siber.FilterInfo) (caseList *[]*siber.CaseInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	t, err := getMongoCollection(ctx, CollectionTag)
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

	// 模糊搜索
	if _, ok := filterInfo.FilterContent["content"]; ok {
		var filter = make(map[string]string)
		filter["casename"] = filterInfo.FilterContent["content"]
		filter["methodname"] = filterInfo.FilterContent["content"]
		tagList := new([]*siber.TagInfo)
		err = t.Find(bson.M{"tagname": filterInfo.FilterContent["content"]}).All(tagList)
		var tagIds []string
		for _, v := range *tagList {
			tagIds = append(tagIds, v.TagId)
		}
		if len(tagIds) > 0 {
			for _, v := range tagIds {
				filter["casetags"] = v
			}
		}
		var orQueries []bson.M
		for k, v := range filter {
			orQueries = append(orQueries, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		var andQuery []bson.M
		andQuery = append(andQuery, bson.M{"$or": orQueries})
		andQuery = append(andQuery, bson.M{"invaliddate": 0})
		query = bson.M{"$and": andQuery}
	}

	// 精准搜索
	if _, ok := filterInfo.FilterContent[MethodNode]; ok {
		andQuery := []bson.M{
			{"methodname": filterInfo.FilterContent[MethodNode]},
			{"invaliddate": 0},
		}
		query = bson.M{"$and": andQuery}
	}

	if query == nil {
		query = bson.M{"invaliddate": 0}
	}
	var cases []*siber.CaseInfo
	err = c.Find(query).Sort("-updatetime").Select(bson.M{"updatetime": -1, "caseid": 1, "casename": 1, "inserttime": -1, "casemode": 1}).Skip((page - 1) * pageSize).Limit(pageSize).All(&cases)
	if err != nil {
		xzap.Logger(ctx).Error("get case list failed, err", zap.Any("err", err))
	}

	totalNum, err = c.Find(query).Count()
	if err != nil {
		xzap.Logger(ctx).Error("get plan total num failed, err", zap.Any("err", err))
		return
	}
	return &cases, totalNum, err
}

func StatCase(ctx context.Context, query bson2.M) (int, error) {
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		xzap.Logger(ctx).Error("get mongo collection error", zap.Any("err", err))
		return 0, err
	}
	num, err := c.Find(query).Count()
	if err != nil {
		xzap.Logger(ctx).Error("find case num from the collection error", zap.Any("err", err))
		return 0, err
	}
	return num, err
}

func StatCaseLog(ctx context.Context, query bson2.M) (int, error) {
	c, err := getMongoCollection(ctx, CollectionLogCase)
	if err != nil {
		xzap.Logger(ctx).Error("get mongo collection error", zap.Any("err", err))
		return 0, err
	}
	num, err := c.Find(query).Count()
	if err != nil {
		xzap.Logger(ctx).Error("find case num from the collection error", zap.Any("err", err))
		return 0, err
	}
	return num, err
}

func (dao *Dao) UpdateCaseTime(ctx context.Context, caseInfoInput *siber.CaseInfo) (err error) {
	if caseInfoInput == nil || caseInfoInput.CaseId == "" {
		return
	}
	c, err := getMongoCollection(ctx, CollectionCase)
	if err != nil {
		return
	}
	_, _ = c.Upsert(bson.M{"caseid": caseInfoInput.CaseId}, bson.M{"$set": caseInfoInput})
	return
}
