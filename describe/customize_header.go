/**
* @Author: TongTongLiu
* @Date: 2021/3/10 9:04 PM
**/

package describe

import (
	"api-test/dao"
	"api-test/payload"
	"api-test/siberconst"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

// 描述自定义算法（通常是鉴权）的 Header，比如：SiberAuth
func DescCustomizeHeader(ctx context.Context, headerOriginal map[string]string, env string) (headers map[string]string, err error) {
	headers = make(map[string]string)
	for k, v := range headerOriginal {
		kValue := siberconst.Siber + "#"
		// 没有使用自定义的 Header
		if !strings.HasPrefix(k, kValue) {
			headers[k] = v
			continue
		}

		// 使用了自定义的 Header
		keyList := strings.Split(k, "#")
		if len(keyList) != 3 {
			errMsg := "不正确的 Header Key： " + k + " ，请检查！"
			xzap.Logger(ctx).Error(errMsg)
			err = status.Errorf(codes.Canceled, errMsg)
			return
		}
		productName := keyList[1]
		secretName := keyList[2]

		switch v {
		case siberconst.SiberAuth:
			var h map[string]string
			h, err = siberAuthHeader(ctx, secretName, productName, env)
			if err != nil || h == nil {
				return
			}
			for kk, vv := range h {
				headers[kk] = vv
			}
		default:
			errMsg := "不支持的自定义算法： " + v + " ，请检查！"
			xzap.Logger(ctx).Info(errMsg)
			err = status.Errorf(codes.Canceled, errMsg)
			return
		}
	}
	return
}

//func descInterfaceRequest(ctx context.Context, caseInfo *siber.CaseInfo, c *Case) (interRequest *api.InterfaceRequest, err error) {
//	headers, err := descCustomizeHeader(ctx, c.Request.Header, c.Plan.Trigger.Protocol)
//	if err != nil {
//		return
//	}
//	interRequest = &api.InterfaceRequest{
//		MethodName: caseInfo.MethodName,
//		Header:     headers,
//		// TODO: body 这里渲染 func variable
//		Body: nil,
//		// TODO: URL 这里根据环境进行渲染
//		URL: "",
//	}
//	return
//}

// 根据环境中配置的 pubKey、secret 和 env渲染可鉴权的 header
//   - productName:用户自定义的项目名称
//   - env: 环境类型：开发、测试、灰度、生产
func siberAuthHeader(ctx context.Context, SecretName string, productName string, env string) (headers map[string]string, err error) {
	var envInfo *siber.EnvInfo
	envInfo, err = dao.NewDao().SelectEnv(ctx, &siber.EnvInfo{
		EnvName: productName,
	})
	if err != nil {
		return
	}
	if envInfo == nil {
		errMsg := "未找到自定义 header 中引用的产品线名称： " + productName + " 环境： " + env + " ，请检查！"
		xzap.Logger(ctx).Info(errMsg)
		err = status.Errorf(codes.Canceled, errMsg)
		return
	}

	var secret, pubKey string
	for _, c := range envInfo.SecretList {
		if c.SecretName == SecretName {
			switch env {
			case siberconst.EnvironmentDev:
				valSecret, ok1 := c.SecretInfo.DevSecret[siberconst.SiberSecret]
				valPubkey, ok2 := c.SecretInfo.DevSecret[siberconst.SiberPubkey]
				if ok1 && ok2 {
					secret = valSecret
					pubKey = valPubkey
				}
			case siberconst.EnvironmentTest:
				valSecret, ok1 := c.SecretInfo.TestSecret[siberconst.SiberSecret]
				valPubkey, ok2 := c.SecretInfo.TestSecret[siberconst.SiberPubkey]
				if ok1 && ok2 {
					secret = valSecret
					pubKey = valPubkey
				}
			case siberconst.EnvironmentStage:
				valSecret, ok1 := c.SecretInfo.StageSecret[siberconst.SiberSecret]
				valPubkey, ok2 := c.SecretInfo.StageSecret[siberconst.SiberPubkey]
				if ok1 && ok2 {
					secret = valSecret
					pubKey = valPubkey
				}
			case siberconst.EnvironmentProd:
				valSecret, ok1 := c.SecretInfo.ProdSecret[siberconst.SiberSecret]
				valPubkey, ok2 := c.SecretInfo.ProdSecret[siberconst.SiberPubkey]
				if ok1 && ok2 {
					secret = valSecret
					pubKey = valPubkey
				}
			}
			break
		}
	}
	if secret == "" {
		errMsg := "自定义 Header：" + env + " ，secret 为空，请检查！"
		xzap.Logger(ctx).Info(errMsg)
		status.Errorf(codes.Canceled, errMsg)
		return
	}
	if pubKey == "" {
		errMsg := "自定义 Header：" + env + " ，pubKey 为空，请检查！"
		xzap.Logger(ctx).Info(errMsg)
		status.Errorf(codes.Canceled, errMsg)
		return
	}

	credential := payload.NewCredential(secret, pubKey)
	headers = make(map[string]string)
	for k, v := range credential.GetHeaders() {
		headers[k] = v
	}
	return
}
