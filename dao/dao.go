/**
* @Author: TongTongLiu
* @Date: 2019-09-14 10:04
**/

package dao

import (
	"api-test/configs"
	"api-test/libs"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"github.com/globalsign/mgo"
	bson2 "github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
)

// siber 的集合
const (
	CollectionPlan           = "collection_plan"
	CollectionFlow           = "collection_flow"
	CollectionCase           = "collection_case"
	CollectionCaseVersion    = "collection_case_version"
	CollectionTag            = "collection_tag"
	CollectionEnv            = "collection_env"
	CollectionMethod         = "collection_method"
	CollectionLogCase        = "collection_log_case"
	CollectionLogFlow        = "collection_log_flow"
	CollectionLogPlan        = "collection_log_plan"
	CollectionRunStatus      = "collection_run_status"
	CollectionProcessPlan    = "collection_process_plan"
	CollectionProcessPlanLog = "collection_process_plan_log"
	CollectionGraphqlQuery   = "collection_graphql_query"
)
// ops 的集合
const CollectionProcesses = "processes"

const (
	PackageNode = "package"
	ServiceNode = "service"
	MethodNode  = "method"
	CaseNode    = "case"
	FlowNode    = "flow"
	PlanNode    = "plan"
)

const defaultPage = 1
const defaultPageSize = 100

const ValidTag = 0

type objectID struct {
	Id bson2.ObjectId `json:"id" bson:"_id"`
}

type Dao struct {
}

func NewDao() *Dao {
	return &Dao{}
}

// siber 的库
func getMongoCollection(ctx context.Context, collectionName string) (c *mgo.Collection, err error) {
	config := configs.GetGlobalConfig().Mongo
	dbName := configs.GetGlobalConfig().Mongo.DBName
	session, err := libs.GetMongoSession(config, dbName)
	if err != nil {
		xzap.Logger(ctx).Error("connect to mongoDB failed", zap.Any("err", err))
		return
	}
	c = session.DB(configs.GetGlobalConfig().Mongo.DBName).C(collectionName)
	return
}

// 外部库
func getOpsCollection(ctx context.Context, collectionName string) (c *mgo.Collection, err error) {
	config := configs.GetGlobalConfig().MongoOps
	dbName := configs.GetGlobalConfig().MongoOps.DBName
	session, err := libs.GetMongoSession(config, dbName)
	if err != nil {
		xzap.Logger(ctx).Error("connect to mongoDB ops failed", zap.Any("err", err))
		return
	}
	c = session.DB(configs.GetGlobalConfig().MongoOps.DBName).C(collectionName)
	return
}
// TODO: 比较鸡肋，可以删除
func isInputNull(ctx context.Context, input interface{}) (err error) {
	if input != nil {
		return
	}
	xzap.Logger(ctx).Info("input is nil")
	err = status.Errorf(codes.InvalidArgument, "input %s is null", reflect.TypeOf(input))
	return
}

// TODO: 用于拼接根据条件查询的内容
//func conditionConcat(ctx context.Context, info *siber.FilterInfo) (cond string, err error) {
//	cond = `"invaliddate": 0`
//	if info == nil {
//		return
//	}
//	if _, ok := info.FilterContent[ApplicationSymbol]; ok {
//		cond = fmt.Sprintf(`%s, "applicationname":"%s"`, cond, info.FilterContent[ApplicationSymbol])
//	}
//	if _, ok := info.FilterContent[ServiceSymbol]; ok {
//		cond = fmt.Sprintf(`%s, "servicename":"%s"`, cond, info.FilterContent[ApplicationSymbol])
//	}
//	return
//}
