/**
* @Author: TongTongLiu
* @Date: 2020/5/15 10:52 上午
**/

package sibercron

import (
	"api-test/dao"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"sort"
)

// 刷新所有的plan 列表，避免非常规手段加的plan无法被触发，或者偶然err 导致的更新遗漏
func RefreshAllPlanServices(ctx context.Context) (err error) {
	// 偷懒的写法，如果将来plan多了，需要分批刷新
	filter := &siber.FilterInfo{
		FilterContent: map[string]string{
			"page":      "1",
			"page_size": "2000",
		},
	}
	planList, _, err := dao.NewDao().ListPlan(ctx, filter)
	if err != nil {
		return
	}
	for _, p := range *planList {
		oServices := p.Services
		err = UpdatePlanServices(ctx, p)
		if err != nil {
			continue
		}
		if stringSliceEqualBCE(oServices, p.Services) {
			continue
		}
		_, _ = dao.NewDao().UpdatePlan(ctx, p)
	}
	xzap.Logger(ctx).Info("RefreshAllPlanServices success")
	return
}

// 更新plan中涉及到的所有service，方便测试环境自动触发plan运行
func UpdatePlanServices(ctx context.Context, planInput *siber.PlanInfo) (err error) {
	if planInput == nil {
		return
	}
	planInput.Services, err = GetPlanServices(ctx, planInput)
	return
}

// 获得plan绑定的相关service
// 调用入口有：定时订正，plan创建及修改
func GetPlanServices(ctx context.Context, planInput *siber.PlanInfo) (serviceList []string, err error) {
	if planInput == nil {
		return
	}
	// plan 中flow不重复，省略去重步骤
	// 获得case 列表
	var caseList []string
	for _, f := range planInput.FlowList {
		flowInfo, err := dao.NewDao().SelectFlow(ctx, &siber.FlowInfo{
			FlowId: f,
		})
		if err != nil || flowInfo == nil {
			continue
		}
		caseList = append(caseList, flowInfo.CaseList...)
	}
	caseList = removeRepByMap(caseList)

	//  从case详情中获得method 列表
	var methodList []string
	for _, c := range caseList {
		caseInfo, err := dao.NewDao().SelectCase(ctx, &siber.CaseInfo{
			CaseId: c,
		})
		if err != nil || caseInfo == nil {
			continue
		}
		if caseInfo.MethodName == "" {
			continue
		}
		methodList = append(methodList, caseInfo.MethodName)
	}
	methodList = removeRepByMap(methodList)

	// 获得 service 列表
	for _, m := range methodList {
		methodInfo, err := dao.NewDao().SelectMethod(ctx, &siber.MethodInfo{
			MethodName: m,
		})
		if err != nil || methodInfo == nil {
			return nil, err
		}
		serviceList = append(serviceList, methodInfo.Service)
	}
	serviceList = removeRepByMap(serviceList)
	return
}

// 列表去重并排序，方便后续比对
// 通过map主键唯一的特性过滤重复元素
func removeRepByMap(slc []string) []string {
	var result []string
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	sort.Sort(sort.StringSlice(result))
	return result
}

func stringSliceEqualBCE(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
