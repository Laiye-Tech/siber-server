/**
* @Author: TongTongLiu
* @Date: 2019-09-12 15:00
**/

package service

import (
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/cibot"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func getClient() (client siber.SiberServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "localhost:8888", grpc.WithInsecure())
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	client = siber.NewSiberServiceClient(conn)
	return
}

func NewCiGetClient() (client ci.CiServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "localhost:8888", grpc.WithInsecure())
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	client = ci.NewCiServiceClient(conn)
	return
}

func TestSiberService_ParseMethodList(t *testing.T) {
	client := getClient()
	resp, err := client.ParseMethodList(context.Background(), &siber.ParseMethodListRequest{

		ProtoFile: "git.laiye.com/laiye-backend-repos/im-saas-protos-golang/manage_user/manage_user.proto",
	})
	fmt.Println(resp, err)
}

//func TestSiberService_GetMethodList(t *testing.T) {
//	client := getClient()
//	m := new(siber.FilterInfo)
//	resp, err := client.GetMethodTree(context.Background(), m)
//	fmt.Println(resp, err)
//}
//
//func TestSiberService_GetTreeStruct(t *testing.T) {
//	client := getClient()
//	m := new(siber.FilterInfo)
//	resp, err := client.GetTreeStruct(context.Background(), m)
//	fmt.Println(resp, err)
//}

func TestSiberService_ManageCase(t *testing.T) {
	client := getClient()
	m := new(siber.ManageCaseInfo)
	m.ManageMode = "CREATE"
	m.CaseInfo = &siber.CaseInfo{
		MethodName: "manage_user.ManageUserService/CreateUser",
		CaseName:   "tina测试插入用户数据",
	}
	resp, err := client.ManageCase(context.Background(), m)
	fmt.Println(resp, err)
}

func TestSiberService_ManageFlow(t *testing.T) {
	client := getClient()
	m := new(siber.ManageFlowInfo)
	m.ManageMode = "CREATE"
	m.FlowInfo = &siber.FlowInfo{
		FlowName: "tinatest flow",
	}
	resp, err := client.ManageFlow(context.Background(), m)
	fmt.Println(resp, err)
}

func TestSiberService_ManagePlan(t *testing.T) {
	client := getClient()
	m := new(siber.ManagePlanInfo)
	m.ManageMode = "CREATE"
	m.PlanInfo = &siber.PlanInfo{
		//PlanId:   "5d87631873c56fecf0117e3b",
		PlanName: "test3",
	}
	resp, err := client.ManagePlan(context.Background(), m)
	fmt.Println(resp, err)
}

func TestSiberService_ListCIProject(t *testing.T) {
	client := NewCiGetClient()
	m := new(ci.GetAllProjectsRequest)

	resp, err := client.GetAllProjects(context.Background(), m)
	fmt.Println(resp, err)
}
