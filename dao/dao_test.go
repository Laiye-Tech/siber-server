/**
* @Author: TongTongLiu
* @Date: 2019-09-17 12:07
**/

package dao

import (
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"reflect"
	"testing"
)

//func TestDao_InsertMethods(t *testing.T) {
//	type args struct {
//		ctx        context.Context
//		methodList *siber.ParseMethodListResponse
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			dao := &Dao{}
//			if err := dao.InsertMethods(tt.args.ctx, tt.args.methodList); (err != nil) != tt.wantErr {
//				t.Errorf("InsertMethods() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestNewDao(t *testing.T) {
	tests := []struct {
		name string
		want *Dao
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDao(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDao() = %v, want %v", got, tt.want)
			}
		})
	}
}

//
//func TestCase_InsertCase(t *testing.T) {
//	caseInput := &siber.ManageCaseInfo{
//		MethodName: "0.0.0",
//		CaseName:   "tinatest",
//	}
//	caseOutPut, err := NewDao().InsertCase(context.Background(), caseInput)
//	fmt.Println("err is:", err)
//	fmt.Println("case output is:", caseOutPut)
//
//}
//
//func TestFlow_InsertFlow(t *testing.T) {
//	flowInput := &siber.ManageFlowInfo{
//		FlowName: "tinatest",
//		Remark:   "test",
//	}
//	caseOutPut, err := NewDao().InsertFlow(context.Background(), flowInput)
//	fmt.Println("err is:", err)
//	fmt.Println("plan output is:", caseOutPut)
//
//}
//
//func TestPlan_InsertPlan(t *testing.T) {
//	planInput := &siber.ManagePlanInfo{
//		PlanName: "tinatest",
//		Remark:   "test",
//	}
//	caseOutPut, err := NewDao().InsertPlan(context.Background(), planInput)
//	fmt.Println("err is:", err)
//	fmt.Println("plan output is:", caseOutPut)
//
//}
//
//func TestMethod_SelectMethod(t *testing.T) {
//	methodList, err := NewDao().SelectMethods(context.Background(), nil)
//	fmt.Println("method list is :", methodList)
//	fmt.Println("err is", err)
//}

func TestPlan_SelectPlan(t *testing.T) {

	m := &siber.PlanInfo{
		PlanId:   "5d885d9973c56fecf0287113",
		PlanName: "testplan2",
	}

	planInfo, err := NewDao().SelectPlan(context.Background(), m)
	fmt.Println("plan info is :", planInfo)
	fmt.Println("err is", err)
}

//func TestMethod_InsertMethod(t *testing.T) {
//	path := configs.GetGlobalConfig().ProtoFile.RootPath
//	importPaths := []string{
//		path,
//		fmt.Sprintf("%s/protos", path),
//		fmt.Sprintf("%s/protos/siber", path),
//		//fmt.Sprintf("%s/protos/saas.openapi.v2", path),
//		//fmt.Sprintf("%s/protos/im_user_attribute", path),
//		//fmt.Sprintf("%s/protos/im_user_attribute", path),
//		//fmt.Sprintf("%s/protos/manage_user", path),
//	}
//	protoFiles := []string{fmt.Sprintf("%s/protos/siber/siber.proto", path)}
//	i := &siber.MethodInfo{
//		MethodName:      "siber.SiberService.ManageCase",
//		ProtoFiles:      protoFiles,
//		ImportPaths:     importPaths,
//		HttpUri:         "/siber/manage/case",
//		HttpRequestMode: "POST",
//	}
//	iout, err := NewDao().InsertInterface(context.Background(), i)
//	fmt.Println("iout is:", iout)
//	fmt.Println("err is :", err)
//}

func TestPlan_ListPlan(t *testing.T) {
	plans, total_num, err := NewDao().ListPlan(context.Background(), nil)
	fmt.Println("plans is", plans)
	fmt.Println("total num is", total_num)
	fmt.Println("err is", err)
}
