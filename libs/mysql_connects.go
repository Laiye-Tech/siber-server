/**
* @Author: TongTongLiu
* @Date: 2020/1/17 3:18 下午
**/

package libs

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

const (
	ConnNum = 20
	TimeOut = 28800
)

type mysqlClient struct {
	Connects map[string]*gorm.DB
}

var client *mysqlClient
var lock = &sync.RWMutex{}

type mysqlLogger struct{}

func (m *mysqlLogger) Print(v ...interface{}) { Log().Info(context.Background(), "%v", v...) }

func connectDbs(config *siber.InstanceType, instanceName string) error {
	if config == nil {
		return status.Errorf(codes.InvalidArgument, "db config is nil")
	}
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		config.User, config.Password, config.Host, config.Port, config.Db, config.Charset)
	o, err := gorm.Open("mysql", connectString)
	if err != nil {
		Log().Error(context.Background(), "connect db(%v) error(%v)", config, err)
		return err
	}
	o.DB().SetMaxIdleConns(ConnNum)
	o.DB().SetMaxOpenConns(ConnNum)
	o.DB().SetConnMaxLifetime(TimeOut * time.Second)
	
	o.LogMode(true)
	o.SetLogger(&mysqlLogger{})
	client.Connects[instanceName] = o
	return nil
}

func GetMysqlDb(config *siber.InstanceType, instanceName string) (o *gorm.DB, err error) {
	if v, ok := client.Connects[instanceName]; ok {
		return v, err
	} else {
		lock.Lock()
		defer lock.Unlock()
		if v, ok := client.Connects[instanceName]; ok {
			return v, err
		}
		err = connectDbs(config, instanceName)
		return client.Connects[instanceName], err
	}
}

func init() {
	client = &mysqlClient{
		Connects: make(map[string]*gorm.DB),
	}
}
