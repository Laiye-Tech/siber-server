package libs

import (
	"api-test/configs"
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"github.com/globalsign/mgo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

var (
	mongoSessions = make(map[string]*mgo.Session)
	mgoLocker     = new(sync.RWMutex)
)

type mongoConfig struct {
	uri           string
	name          string
	password      string
	db            string
	host          string
	port          int
	maxPoolSize   int
	minPoolSize   int
	maxIdleTimeMs int
}

func getMongoConfig(config configs.Mongo, dbName string) (*mongoConfig, error) {
	host := config.Host
	if host == "" {
		return nil, status.Errorf(codes.NotFound, "mongo instance %v not exist", config.Host)
	}
	port := config.Port
	name := config.Name
	password := config.Password
	db := dbName
	maxPoolSize := config.MaxPoolSize
	minPoolSize := config.MinPoolSize
	maxIdleTimeS := config.MaxIdleTime
	uri := config.Uri
	return &mongoConfig{
		uri:           uri,
		name:          name,
		password:      password,
		db:            db,
		host:          host,
		port:          port,
		maxPoolSize:   maxPoolSize,
		minPoolSize:   minPoolSize,
		maxIdleTimeMs: maxIdleTimeS * 1000,
	}, nil
}

func mongoKeyName(config configs.Mongo, dbName string) string {
	return fmt.Sprintf("%v:%v", config.Name, dbName)
}

func GetMongoSession(config configs.Mongo, dbName string) (*mgo.Session, error) {
	ctx := context.Background()
	var s *mgo.Session
	mgoLocker.RLock()
	key := mongoKeyName(config, dbName)
	s = mongoSessions[key]
	mgoLocker.RUnlock()
	if s == nil {
		cfg, err := getMongoConfig(config, dbName)
		if err != nil {
			return nil, err
		}
		//url := fmt.Sprintf("mongodb://%v:%v@%v:%v/%v?maxPoolSize=%v&"+
		//	"minPoolSize=%v&maxIdleTimeMS=%v&authSource=%v", cfg.name, cfg.password, cfg.host, cfg.port,
		//	cfg.db, cfg.maxPoolSize, cfg.minPoolSize, cfg.maxIdleTimeMs, cfg.db)
		url := fmt.Sprintf("%v&maxPoolSize=%v&minPoolSize=%v&maxIdleTimeMS=%v&authSource=%v",
			cfg.uri, cfg.maxPoolSize, cfg.minPoolSize, cfg.maxIdleTimeMs, cfg.db)
		st, err := mgo.Dial(url)
		xzap.Logger(ctx).Debug("connect url", zap.Any("url", url))

		if err != nil {
			xzap.Logger(ctx).Error("connect mongo failed", zap.Any("mongo", key), zap.Any("err", err))
			return nil, err
		} else {
			xzap.Logger(ctx).Info("connect mongo success", zap.Any("key", key))
		}
		s = st
		mgoLocker.Lock()
		mongoSessions[key] = s
		mgoLocker.Unlock()
	}
	return s, nil
}
