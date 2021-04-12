package api

import (
	"api-test/libs"
	"api-test/payload"
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

const (
	Siber       = "Siber"
	SiberAuth   = "SiberAuth"
	SiberPubkey = "pubkey"
	SiberSecret = "secret"
)

const (
	verboseResponseHeader = `
Response headers received:
`
	verboseResponseContents = `
Response contents:
`
	verboseResponseTrailer = `
Response trailers received:
`
)

type Interface interface {
	Invoke(ctx context.Context, request *payload.Request, environment string) (pResp *payload.Response, err error)
	Url() string
	Uri() string
}

func (r *HTTPInterface) Url() string {
	return r.ReqPath
}

func (g *GRPCInterface) Url() string {
	return g.ReqPath
}

func (r *HTTPInterface) Uri() string {
	return ""

}

func ParseUrl(r *HTTPInterface, request *payload.Request) (parseReq *url.URL, err error) {
	urlParse, err := url.Parse(r.ReqPath)
	if err != nil {
		return nil, err
	}
	q, err := url.ParseQuery(urlParse.RawQuery)
	if err != nil {
		return nil, err
	}
	// 渲染body到urlParameter
	if r.ReqMode == "GET" && len(request.Body) != 0 {
		var data map[string]interface{}
		err := json.NewDecoder(strings.NewReader(string(request.Body))).Decode(&data)
		if err != nil {
			return nil, err
		}
		for k, v := range data {
			q.Add(k, libs.ToStr(v))
		}
	}
	parameterParse, _ := url.Parse(request.UrlParameter)
	// 如果UrlParameter 参数传递方式
	if parameterParse.RawQuery == "" && request.UrlParameter != "" {
		parseResult, err := url.ParseQuery(request.UrlParameter)
		if err != nil {
			return nil, err
		}
		for i, j := range parseResult {
			for _, n := range j {
				q.Add(i, n)
			}
		}
	}
	// 如果UrlParameter 是一个完整的路径
	if parameterParse.RawQuery != "" && request.UrlParameter != "" {
		parseResult, err := url.ParseQuery(parameterParse.RawQuery)
		if err != nil {
			return nil, err
		}
		for i, j := range parseResult {
			for _, n := range j {
				q.Add(i, n)
			}
		}
	}
	urlParse.Path += parameterParse.Path
	urlParse.RawQuery = q.Encode()
	return urlParse, nil

}
