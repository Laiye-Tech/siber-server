package payload

import (
	"bytes"
	"errors"
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
)

const bodyJson = "$json"

type Selector interface {
	Render(ctx context.Context, v *Variable) (err error)
	Create()
	Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error)
}

type ResponseBodySelector struct {
	Selector `json:"Selector,omitempty"`
	Name     string
	Key      string
	Value    interface{}
}

type RequestBodySelector struct {
	Selector `json:"Selector,omitempty"`
	Name     string
	Key      string
	Value    interface{}
}

type ResponseHeaderSelector struct {
	Selector `json:"Selector,omitempty"`
	Name     string
	Key      string
}

type RequestHeaderSelector struct {
	Selector `json:"Selector,omitempty"`
	Name     string
	Key      string
}

type CostTimeSelector struct {
	Selector `json:"Selector,omitempty"`
	Name     string
}

type StatusCodeSelector struct {
	Selector `json:"Selector,omitempty"`
	Name     string
}

func bodySelector(ctx context.Context, body []byte, selectorKey string) (keyExits bool, value interface{}, err error) {
	bodyString := string(body)
	// 用户想要返回全部 response body
	if selectorKey == bodyJson {
		value = string(body)
		return
	}
	// 从json中截取相应结果
	if !gjson.Valid(bodyString) {
		return false, nil, status.Errorf(codes.InvalidArgument, "invalid json")
	} else {
		selectorKeyParse := strings.Split(selectorKey, "|@sub")
		data := gjson.Get(bodyString, selectorKeyParse[0])
		if !data.Exists() {
			return false, nil, nil
		}
		if len(selectorKeyParse) > 2 {
			return false, nil, status.Errorf(codes.InvalidArgument, "select err value %v", selectorKey)

		}
		if len(selectorKeyParse) == 1 {
			return true, data.Value(), nil

		} else {
			strValue, ok := data.Value().(string)
			if !ok {
				return true, data.Value(), nil
			}
			value := match(selectorKeyParse[1])
			if value != "" {
				num, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return false, nil, status.Errorf(codes.InvalidArgument, "select err value %v", selectorKey)
				}
				return true, strValue[0:num], nil
			}
			return true, data.Value(), nil
		}
	}
}

func renderSelector(ctx context.Context, key string, v *Variable) (renderKey string, err error) {
	parametersName, err := ExtractParameter(ctx, key, v)
	if err != nil {
		return renderKey, err
	}
	if len(parametersName) == 0 {
		renderKey = key
		return renderKey, nil
	}
	for k, v := range parametersName {
		strV, ok := (v).(string)
		if !ok {
			return renderKey, err
		}
		if checkParameter(k) {
			renderKey = strV
		} else {
			nameByte := bytes.ReplaceAll([]byte(strV), []byte(k), []byte(fmt.Sprintf("%v", v)))
			renderKey = string(nameByte)
		}
	}
	return renderKey, nil
}

func (selector *ResponseBodySelector) Render(ctx context.Context, v *Variable) (err error) {
	renderKey, err := renderSelector(ctx, selector.Key, v)
	if err != nil {
		return
	}
	selector.Key = renderKey
	return
}
func (selector *RequestBodySelector) Render(ctx context.Context, v *Variable) (err error) {
	renderKey, err := renderSelector(ctx, selector.Key, v)
	if err != nil {
		return
	}
	selector.Key = renderKey
	return
}

func (selector *RequestHeaderSelector) Render(ctx context.Context, v *Variable) (err error) {
	renderKey, err := renderSelector(ctx, selector.Key, v)
	if err != nil {
		return
	}
	selector.Key = renderKey
	return
}

func (selector *ResponseHeaderSelector) Render(ctx context.Context, v *Variable) (err error) {
	renderKey, err := renderSelector(ctx, selector.Key, v)
	if err != nil {
		return
	}
	selector.Key = renderKey
	return
}

func (costTimeSelector *CostTimeSelector) Render(ctx context.Context, v *Variable) (err error) {
	return
}
func (statusSelector *StatusCodeSelector) Render(ctx context.Context, v *Variable) (err error) {
	return
}

func (selector *ResponseBodySelector) Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error) {
	if selector == nil {
		xzap.Logger(ctx).Info("Selector got nil selector")
		err = status.Errorf(codes.InvalidArgument, "Select got nil selector")
		return
	}
	if response == nil {
		xzap.Logger(ctx).Info("Selector got nil response")
		err = status.Errorf(codes.InvalidArgument, "Selector got nil response")
		return
	}

	keyExits, value, err = bodySelector(ctx, response.Body, selector.Key)
	selector.Value = value
	return
}

func (selector *RequestBodySelector) Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error) {
	if selector == nil {
		err = status.Errorf(codes.InvalidArgument, "Select got nil selector")
		return
	}
	if request == nil {
		err = status.Errorf(codes.InvalidArgument, "Selector got nil request")
		return
	}
	keyExits, value, err = bodySelector(ctx, request.Body, selector.Key)
	selector.Value = value
	return
}

func headerSelector(ctx context.Context, header map[string]string, selectorKey string) (keyExits bool, value interface{}, err error) {
	headerValue, ok := header[selectorKey]
	if ok {
		return true, headerValue, nil
	} else {
		return false, nil, errors.New("headerKey does not exist")
	}
}

func (selector *ResponseHeaderSelector) Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error) {
	if request == nil || selector == nil {
		err = status.Errorf(codes.InvalidArgument, "ResponseHeaderSelector Select failed, nil pointer")
		return
	}
	keyExits, value, err = headerSelector(ctx, response.Header, selector.Key)
	return
}

func (selector *RequestHeaderSelector) Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error) {
	if request == nil || selector == nil {
		err = status.Errorf(codes.InvalidArgument, "RequestHeaderSelector Select failed, nil pointer")
		return
	}
	keyExits, value, err = headerSelector(ctx, request.Header, selector.Key)
	return
}

func (statusSelector *StatusCodeSelector) Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error) {
	if response == nil {
		err = status.Errorf(codes.InvalidArgument, "statusSelector Select failed, nil pointer")
		return
	}
	statusCode := strconv.Itoa(response.StatusCode)
	if statusCode != "" {
		return true, response.StatusCode, nil
	} else {
		return true, nil, errors.New("statusCode is null")
	}
}

func (costTimeSelector *CostTimeSelector) Select(ctx context.Context, request *Request, response *Response) (keyExits bool, value interface{}, err error) {
	if response == nil {
		err = status.Errorf(codes.InvalidArgument, "costTimeSelector Select failed, nil pointer")
		return
	}
	costTime := strconv.Itoa(response.TimeCost)
	if costTime != "" {
		return true, response.TimeCost, nil
	} else {
		return true, nil, errors.New("costTime is null")
	}
}
