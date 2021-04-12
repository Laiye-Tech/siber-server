/**
* @Author: TongTongLiu
* @Date: 2021/3/10 11:42 AM
**/

package api

import (
	"api-test/payload"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type HTTPInterface struct {
	// interface
	Interface
	InterfaceNew

	// struct
	InterfaceRequest

	MethodName string

	ReqPath   string
	ReqMode   string
	ReqHeader map[string]string
}

func (r *HTTPInterface) Invoke(ctx context.Context, request *payload.Request, environment string) (pResp *payload.Response, err error) {
	client := &http.Client{}
	parseUrlResult, err := ParseUrl(r, request)
	if err != nil {
		xzap.Logger(ctx).Info("Parse Request url failed, err ", zap.Any("err", err))
		err = status.Errorf(codes.Unknown, "Parse Request url failed,err :%v", err)
		return
	}
	request.UrlParameter = parseUrlResult.RawQuery
	r.ReqPath = parseUrlResult.String()
	req, err := http.NewRequest(r.ReqMode, r.ReqPath, strings.NewReader(string(request.Body)))
	if err != nil {
		xzap.Logger(ctx).Info("New Request failed, err ", zap.Any("err", err))
		err = status.Errorf(codes.Unavailable, "New Request failed, err :%v", err)
		return
	}

	if request.Header != nil && len(request.Header) > 0 {
		for k, v := range request.Header {
			req.Header.Set(k, v)
		}
	}
	beginTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		xzap.Logger(ctx).Info("do restful invoke failed", zap.Any("err", err))
		err = status.Errorf(codes.Unavailable, "do restful invoke failed, err :%v", err)
		return
	}
	endTime := time.Now()
	defer func() {
		_ = resp.Body.Close()
		if err != nil {
			xzap.Logger(ctx).Error("close request failed",
				zap.Any("err", err))
			return
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		xzap.Logger(ctx).Error("read response body failed",
			zap.Any("err", err))
		return
	}

	pRespHeader := make(map[string]string)
	for k, v := range resp.Header {
		pRespHeader[k] = v[0]
	}
	pResp = new(payload.Response)
	pResp.Header = pRespHeader
	pResp.Body = body
	pResp.StatusCode = resp.StatusCode
	pResp.TimeCost = int(endTime.Sub(beginTime).Nanoseconds() / 1e6)
	return
}
