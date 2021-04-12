/**
* @Author: TongTongLiu
* @Date: 2019/10/12 6:38 下午
**/

package dao

import (
	"context"
	"strconv"
	"time"

	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/globalsign/mgo"
	bson2 "github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (dao *Dao) InsertMethod(ctx context.Context, methodInput *siber.MethodInfo) (methodOutput *siber.MethodInfo, err error) {
	c, err := getMongoCollection(ctx, CollectionMethod)
	if err != nil {
		return
	}
	methodInput.InsertTime = time.Now().Unix()
	methodInput.UpdateTime = time.Now().Unix()
	err = c.Insert(methodInput)
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "method name is a duplicate ")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to Insert method ")
		return
	}
	methodOutput, err = dao.SelectMethod(ctx, methodInput)
	return
}

func (dao *Dao) SelectMethod(ctx context.Context, methodInput *siber.MethodInfo) (methodOutput *siber.MethodInfo, err error) {
	err = isInputNull(ctx, methodInput)
	if err != nil {
		return
	}
	if methodInput.MethodName == "" {
		err = status.Errorf(codes.InvalidArgument, "methodName is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionMethod)
	if err != nil {
		return
	}
	methodOutput = new(siber.MethodInfo)
	err = c.Find(bson.M{"methodname": methodInput.MethodName}).One(&methodOutput)
	if err != nil {
		err = status.Errorf(codes.Aborted, "failed to persistence interface info , err: %v", err)
	}
	return
}

func (dao *Dao) UpdateMethod(ctx context.Context, methodInput *siber.MethodInfo) (methodOutput *siber.MethodInfo, err error) {
	err = isInputNull(ctx, methodInput)
	if err != nil {
		return
	}
	if methodInput.MethodName == "" {
		err = status.Errorf(codes.InvalidArgument, "methodName is null")
		return
	}
	c, err := getMongoCollection(ctx, CollectionMethod)
	if err != nil {
		return
	}
	methodInput.UpdateTime = time.Now().Unix()
	err = c.Update(bson.M{"methodname": methodInput.MethodName}, bson.M{"$set": methodInput})
	if err != nil {
		if mgo.IsDup(err) {
			err = status.Errorf(codes.AlreadyExists, "method name is a duplicate ")
			return
		}
		err = status.Errorf(codes.InvalidArgument, "failed to update Method, err: %v", err)
		return
	}
	methodOutput, err = dao.SelectMethod(ctx, methodInput)
	return
}

func (dao *Dao) ListMethod(ctx context.Context, filterInfo *siber.FilterInfo) (MethodList *[]*siber.MethodInfo, totalNum int, err error) {
	c, err := getMongoCollection(ctx, CollectionMethod)
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
		filter["methodname"] = filterInfo.FilterContent["content"]
		filter["httpuri"] = filterInfo.FilterContent["content"]
		filter["httprequestmode"] = filterInfo.FilterContent["content"]
		var orQueries []bson.M
		for k, v := range filter {
			orQueries = append(orQueries, bson.M{k: bson.M{"$regex": v, "$options": "$i"}})
		}
		query = bson.M{"$or": orQueries}
	}

	// 精准匹配，用于星图
	if _, ok := filterInfo.FilterContent[ServiceNode]; ok {
		methodName := bson2.RegEx{filterInfo.FilterContent[ServiceNode] + ".*", ""}
		query = bson.M{"methodname": methodName}
	}

	var Methods []*siber.MethodInfo
	err = c.Find(query).Sort("-updatetime").Skip((page - 1) * pageSize).Limit(pageSize).All(&Methods)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get Method list failed, err : %v", err)
		return
	}
	totalNum, err = c.Find(query).Count()
	if err != nil {
		err = status.Errorf(codes.Aborted, "get plan total num failed, err : %v", err)
		return
	}
	return &Methods, totalNum, err
}
func (dao *Dao) ListGraphqlMethod(ctx context.Context, GraphqlRequest *siber.GraphqlMethodListRequest) (GraphqlResponse *siber.GraphqlMethodListResponse, err error) {
	c, err := getMongoCollection(ctx, CollectionGraphqlQuery)
	if err != nil {
		return
	}
	var GraphqlResult []string
	GraphqlResponse = new(siber.GraphqlMethodListResponse)
	err = c.Find(nil).Distinct("methodname", &GraphqlResult)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get graphql list failed, err : %v", err)
		return
	}
	GraphqlResponse.GraphqlMethods = GraphqlResult
	return
}

func (dao *Dao) GetGraphqlQuery(ctx context.Context, methodInput *siber.MethodInfo) (methodOutput *siber.MethodInfo, err error) {
	c, err := getMongoCollection(ctx, CollectionGraphqlQuery)
	if err != nil {
		return
	}
	var uriInfo *struct {
		Uri string `json:"uri" bson:"uri"`
	}
	var graphqlInfo []*siber.GraphqlQueryDetail
	var methodInfo siber.MethodInfo
	err = c.Find(bson.M{"methodname": methodInput.MethodName}).Sort("-version").All(&graphqlInfo)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get graphql Query failed, err : %v", err)
		return
	}
	err = c.Find(bson.M{"methodname": methodInput.MethodName}).One(&uriInfo)
	if err != nil {
		err = status.Errorf(codes.Aborted, "get graphql uri failed, err : %v", err)
		return
	}
	methodInfo.GraphqlQueryDetail = graphqlInfo
	methodInfo.HttpUri = uriInfo.Uri
	return &methodInfo, err
}
