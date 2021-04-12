/**
* @Author: TongTongLiu
* @Date: 2020/4/20 5:00 下午
**/

package api

import (
	"api-test/dao"
	"api-test/sibercron"
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"strconv"
	"strings"
)

const (
	pageNumValue  = 1
	pageSizeValue = 10000
	Page          = "page"
	PageSize      = "page_size"
)

func GetPackageList(ctx context.Context, request *siber.FilterInfo) (response *siber.PackageList, err error) {
	if request == nil {
		request = new(siber.FilterInfo)
	}
	if request.FilterContent == nil {
		request.FilterContent = make(map[string]string)
	}
	request.FilterContent[Page] = strconv.Itoa(pageNumValue)
	request.FilterContent[PageSize] = strconv.Itoa(pageSizeValue)
	methodList, _, err := dao.NewDao().ListMethod(ctx, request)
	if err != nil {
		return
	}

	response = new(siber.PackageList)
	packageMap := make(map[string]string)
	packageList := []string{}
	for _, m := range *methodList {
		packageName := strings.Split(m.MethodName, ".")[0]
		packageMap[packageName] = ""
	}
	for k, _ := range packageMap {
		packageList = append(packageList, k)
	}
	response.PackageName = packageList
	return
}

func GetServiceList(ctx context.Context, request *siber.FilterInfo) (response *siber.ServiceList, err error) {
	if request == nil {
		request = new(siber.FilterInfo)
	}
	if request.FilterContent == nil {
		request.FilterContent = make(map[string]string)
	}
	request.FilterContent[Page] = strconv.Itoa(pageNumValue)
	request.FilterContent[PageSize] = strconv.Itoa(pageSizeValue)
	methodList, _, err := dao.NewDao().ListMethod(ctx, request)
	if err != nil {
		return
	}

	response = new(siber.ServiceList)
	serviceMap := make(map[string]string)
	var serviceList []string
	for _, m := range *methodList {
		method := strings.Split(m.MethodName, ".")
		packageName := method[0]

		// 根据filter中的package筛选
		if request.FilterContent[dao.PackageNode] != "" && request.FilterContent[dao.PackageNode] != method[0] {
			continue
		}

		// 正常情况都是三级，如果只有一级，认为它是package，不在service中显示
		if len(method) < 2 {
			continue
		}
		serviceName := packageName + "." + method[1]
		serviceMap[serviceName] = ""
	}

	for k, _ := range serviceMap {
		serviceList = append(serviceList, k)
	}

	response.ServiceName = serviceList
	return
}

func GetProcessList(request *siber.ProcessListRequest) (response *siber.ProcessListResponse, err error) {
	var processList []string
	response = new(siber.ProcessListResponse)
	for _, m := range *sibercron.Services {
		if m.Service != "" {
			processList = append(processList, m.Service)
		}
	}
	response.ProcessName = processList
	return
}
