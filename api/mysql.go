/**
* @Author: TongTongLiu
* @Date: 2020/1/16 8:13 下午
**/

package api

import (
	"api-test/libs"
	"api-test/payload"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const MaxRows = 100

type Instance interface {
	Init(ctx context.Context, instance *siber.InstanceType) (err error)
	Execute(ctx context.Context, request *payload.Request) (pResp *payload.Response, err error)
	GetInfo(ctx context.Context) (instanceInfo string, err error)
}

type InstanceResult struct {
	Result []map[string]interface{}
}
type MysqlClient struct {
	Instance
	InstanceName string
	InstanceInfo string
	Client       *gorm.DB
}

func (m *MysqlClient) Init(ctx context.Context, instance *siber.InstanceType) (err error) {
	if instance == nil {
		err = status.Errorf(codes.FailedPrecondition, "init mysql client failed, instance is nil")
		return
	}
	m.InstanceInfo = fmt.Sprintf("%s@%s:%d", instance.User, instance.Host, instance.Port)
	m.Client, err = libs.GetMysqlDb(instance, m.InstanceName)
	if err == nil && m.Client == nil {
		err = status.Errorf(codes.FailedPrecondition, "mysql client got nil")
		return
	}
	return
}

// 将查询结果以[]map[string]interface{}形式返回
func getSQLJson(ctx context.Context, rows *sql.Rows) (tableData []map[string]interface{}, err error) {
	columns, err := rows.Columns()
	if err != nil {
		return
	}
	count := len(columns)
	tableData = make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		err = rows.Scan(valuePtrs...)
		if err != nil {
			return
		}
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	return
}

func (m *MysqlClient) Execute(ctx context.Context, request *payload.Request) (pResp *payload.Response, err error) {
	// 这里还可以扩展，比如支持多条语句等
	sql := string(request.Body)
	sql = strings.TrimRight(sql, " ")
	sql = strings.TrimRight(sql, ";")
	sql = fmt.Sprintf("select * from (%s) as siberTmp limit %d;", sql, MaxRows)
	rows, err := m.Client.Raw(sql).Rows()
	if err != nil || rows == nil {
		return
	}
	defer rows.Close()
	sqlResults, err := getSQLJson(ctx, rows)
	if err != nil {
		return
	}
	// 存到标准化json中，方便check & select 结果
	structResult := &InstanceResult{Result: sqlResults}
	pResp = new(payload.Response)
	pResp.Body, err = json.Marshal(structResult)
	return
}

func (m *MysqlClient) GetInfo(ctx context.Context) (instanceInfo string, err error) {
	if m == nil {
		err = status.Errorf(codes.FailedPrecondition, "get mysql instance info failed, client is nil")
		return
	}
	instanceInfo = m.InstanceInfo
	return
}
