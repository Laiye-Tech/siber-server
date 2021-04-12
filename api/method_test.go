/**
* @Author: TongTongLiu
* @Date: 2019-09-11 15:13
**/

package api

import (
	"api-test/configs"
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func Test_getServiceFromProtoFile(t *testing.T) {

	serviceList, err := ParseServiceFromProtoFile(context.Background(), "git.laiye.com/laiye-backend-repos/im-saas-protos-golang/msgId/msgId.proto")
	if err != nil {
		fmt.Printf("test list service failed %+v", err)
	}
	fmt.Println(serviceList)

}

func Test_getMethodsFromProtoFile(t *testing.T) {

	methodList, err := ParseMethodsFromProtoFile(context.Background(), "git.laiye.com/laiye-backend-repos/im-saas-protos-golang/manage_user/manage_user.proto", []string{})
	if err != nil {
		fmt.Printf("test list method failed %+v", err)
	}
	fmt.Println(methodList)
}

//func TestGetMethodTreeInfo(t *testing.T) {
//	resp, err := GetMethodTreeInfo(context.Background(), nil)
//	fmt.Println("resp is :", resp)
//	fmt.Println("err is :", err)
//
//}

func Test_DescribeMethodFromProtoFile(t *testing.T) {
	info := siber.MethodInfo{
		MethodName: "manage_user.ManageUserService.CreateUser",
		ProtoFiles: []string{"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/manage_user/manage_user.proto"},
	}
	resp, err := DescribeMethodFromProtoFile(context.Background(), &info)

	fmt.Println(resp)
	fmt.Println(err)
}
func Test_protoFile(t *testing.T) {
	var files []string
	protoFileRootPath := configs.GetGlobalConfig().ProtoFile.RootPath
	protoFilePath := path.Join(protoFileRootPath, "protos")
	err := filepath.Walk(protoFilePath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".proto") {
			partPath, _ := filepath.Rel(protoFileRootPath, path)
			files = append(files, partPath)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Print(files)
}
