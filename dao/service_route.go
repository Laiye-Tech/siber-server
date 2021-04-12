/**
* @Author: TongTongLiu
* @Date: 2020/5/15 12:32 下午
**/

package dao

import (
	"context"
	"github.com/globalsign/mgo/bson"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
)

// TODO: collection 结构不好，最好改由接口获取
// 据沟通 grpc 和 http 都是唯一的

type RouteUri struct {
	Prefix string `bson:"prefix"`
	Regex  string `bson:"regex"`
}

type Match struct {
	Uri RouteUri `bson:"uri"`
}

type Route struct {
	Matches  []Match `bson:"match"`
	Protocol string  `bson:"protocol"`
	Domain   string  `bson:"domain"`
}
type ServiceRoutes struct {
	Service string  `bson:"process_name"`
	Routes  []Route `bson:"routes"`
}

func (dao *Dao) ListServiceRoute(ctx context.Context) (services *[]*ServiceRoutes, err error) {
	c, err := getOpsCollection(ctx, CollectionProcesses)
	if err != nil {
		return
	}
	var s []*ServiceRoutes
	query := bson.M{"routes": bson.M{"$exists": true}}
	err = c.Find(query).Select(bson.M{"process_name": 1, "routes": 1}).All(&s)
	if err != nil {
		xzap.Logger(ctx).Error("ListServiceRoute failed", zap.Any("err", err))
	}
	services = &s
	return
}
