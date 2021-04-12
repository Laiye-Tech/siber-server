/**
* @Author: TongTongLiu
* @Date: 2020/5/15 11:24 上午
**/

package sibercron

import (
	"api-test/dao"
	"context"
)

// 定时去查service routes相关信息，放到缓存中
// 直接查表不好，太受结构限制，最好能走接口

var Services *[]*dao.ServiceRoutes

// 触发场景：项目启动时set， 定时set，接口触发set
func SetServiceRoutes(ctx context.Context) {
	services, err := dao.NewDao().ListServiceRoute(ctx)
	if err != nil || services == nil {
		return
	}
	Services = services
	return
}
