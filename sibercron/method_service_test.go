/**
* @Author: TongTongLiu
* @Date: 2020/5/18 4:42 下午
**/

package sibercron

import (
	"api-test/initial"
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"testing"
)

func TestGetServicesForMethod(t *testing.T) {
	initial.Initial()
	SetServiceRoutes(context.Background())

	//info := &siber.MethodInfo{
	//	HttpUri:              "/v2/user/create",
	//	MethodType:           "http",
	//}

	info := &siber.MethodInfo{
		MethodName: "manage_user.ManageUserService.CreateUser",
		MethodType: "grpc",
	}
	service, err := GetServicesForMethod(context.Background(), info)
	fmt.Println("service", service)
	fmt.Println("err", err)

}

func TestRefreshAllMethodServices(t *testing.T) {
	initial.Initial()
	SetServiceRoutes(context.Background())
	s := Services
	fmt.Println(s)
 	RefreshAllMethodServices(context.Background())
}
