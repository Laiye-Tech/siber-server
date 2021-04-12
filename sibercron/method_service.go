/**
* @Author: TongTongLiu
* @Date: 2020/5/18 3:39 下午
**/

package sibercron

import (
	"api-test/dao"
	"api-test/siberconst"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

// grpc 是package/service, http 是uri
type uriMatch interface {
	match(ctx context.Context, str string) (match bool, err error)
}

type prefixRoute struct {
	uriMatch
	prefix       string
	protocolType string
}

type regexRoute struct {
	uriMatch
	regex        string
	protocolType string
}

// 刷新所有的method列表，避免非常规手段加的method无法被触发，或者测试同学预定义的method 当时没有配置路由规则
func RefreshAllMethodServices(ctx context.Context) (err error) {
	// 偷懒的写法，如果将来method多了，需要分批刷新
	filter := &siber.FilterInfo{
		FilterContent: map[string]string{
			"page":      "1",
			"page_size": "2000",
		},
	}
	methodList, _, err := dao.NewDao().ListMethod(ctx, filter)
	if err != nil {
		return
	}
	for _, m := range *methodList {
		var service string
		service, err = GetServicesForMethod(ctx, m)
		if err != nil {
			continue
		}
		if m.Service != service && service != "" {
			m.Service = service
			_, err = dao.NewDao().UpdateMethod(ctx, m)
		}
	}
	xzap.Logger(ctx).Info("RefreshAllPlanServices finish")
	return
}

// 找到method对应的service
func GetServicesForMethod(ctx context.Context, info *siber.MethodInfo) (service string, err error) {
	if info == nil {
		return
	}
	if Services == nil {
		xzap.Logger(ctx).Error("Services is nil")
		return
	}
	for _, s := range *Services {
		for _, r := range s.Routes {
			// 域名规则太乱了，当前仅对内部服务做处理
			if strings.Trim(r.Domain, " ") != "*" {
				continue
			}
			// TODO:优化，看能不能用上倒排索引
			var match bool
			match, err = serviceMatch(ctx, &r, info)
			if err != nil {
				return
			}
			if match {
				service = s.Service
				return
			}
		}
	}
	return
}

// 判断指定service 和 method info 是否匹配
// 匹配缓存中存的service-route信息:

// 一个method、uri只能指向唯一一个服务
func serviceMatch(ctx context.Context, route *dao.Route, info *siber.MethodInfo) (match bool, err error) {
	switch info.MethodType {
	case siberconst.HTTPMethod, siberconst.GraphQLMethod:
		if route.Protocol == siberconst.HTTPProtocol {
			match, err = matchUri(ctx, route, info.HttpUri)
			return
		}
	case siberconst.GRPCMethod:
		if route.Protocol == siberconst.GRPCProtocol {
			match, err = matchUri(ctx, route, info.MethodName)
			return
		}

	default:
		xzap.Logger(ctx).Error("GetServicesForMethod failed, unknown method type", zap.Any("type", info.MethodType))
		return
	}

	return
}

// 有前缀和正则两种匹配方式
// 一个规则里面仅同时存在一个
func matchUri(ctx context.Context, route *dao.Route, str string) (match bool, err error) {
	var matcher uriMatch
	for _, m := range route.Matches {
		if m.Uri.Prefix != "" {
			matcher = &prefixRoute{
				prefix:       m.Uri.Prefix,
				protocolType: route.Protocol,
			}

			break
		}
		if m.Uri.Regex != "" {
			matcher = &regexRoute{
				regex:        m.Uri.Regex,
				protocolType: route.Protocol,
			}
			break
		}
	}
	if matcher == nil {
		return
	}
	match, err = matcher.match(ctx, str)
	return
}

/* grpc 类型 route 存储示例
"match": [
    {
        "uri": {
            "prefix": "/saas.openapi.v2.JSSDKSignatureService/"
        }
    }
],
"protocol": "grpc"

*/
func (p *prefixRoute) match(ctx context.Context, str string) (match bool, err error) {
	var prefixStr string
	if p.protocolType == siberconst.GRPCProtocol {
		prefixStr = strings.Trim(p.prefix, "/")
	} else {
		prefixStr = p.prefix
	}

	if prefixStr == "" || str == "" {
		return
	}
	match = strings.HasPrefix(str, prefixStr)
	return
}

func (r *regexRoute) match(ctx context.Context, str string) (match bool, err error) {
	var regexStr string
	if r.protocolType == siberconst.GRPCProtocol {
		regexStr = strings.Trim(r.regex, "/")
	} else {
		regexStr = r.regex
	}
	if regexStr == "" || str == "" {
		return
	}
	match, err = regexp.MatchString(regexStr, str)
	return
}
