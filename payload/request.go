package payload

import (
	"api-test/libs"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"strconv"
	"strings"
)

type Request struct {
	Payload

	parameters []*parameter
}

func CreateRequest() *Request {
	return &Request{}
}
func parameterToStr(ctx context.Context, v interface{}) (value string, err error) {
	switch v.(type) {
	case float64:
		floatV := v.(float64)
		value = strconv.FormatFloat(floatV, 'f', -1, 64)
	case []interface{}, nil:
		ByteV, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		value = string(ByteV)
	default:
		value = fmt.Sprintf("%v", v)
	}
	return
}
func parameterTypeof(ctx context.Context, v interface{}) (value interface{}, err error) {
	if v == nil {
		return v, err
	}
	switch v.(type) {
	case string:
		value = v.(string)
		return
	case int64:
		value = v.(int64)
		return
	case float64:
		value = v.(float64)
		return
	case bool:
		value = v.(bool)
		return
	case []interface{}:
		listV := v.([]interface{})
		if len(listV) == 0 {
			value = make([]string, 0)
			return
		}
		var value []interface{}
		for _, k := range listV {
			vs, err := parameterTypeof(ctx, k)
			if err != nil {
				return nil, err
			}
			value = append(value, vs)
		}
		return value, nil
	default:
		err = status.Errorf(codes.InvalidArgument, "unsupported raw value type, %v", reflect.TypeOf(v))
		return
	}
}

func (r *Request) Render(ctx context.Context, variable *Variable) (err error) {
	str := string(r.Body)
	str += r.UrlParameter
	for _, headerV := range r.Header {
		str += headerV
	}
	parameters, err := ExtractParameter(ctx, str, variable)
	if err != nil {
		return
	}
	for k, v := range parameters {
		var vv string
		v, err = parameterTypeof(ctx, v)
		if err != nil {
			return
		}
		vv, err = parameterToStr(ctx, v)
		if err != nil {
			return
		}
		r.UrlParameter = string(bytes.ReplaceAll([]byte(r.UrlParameter), []byte(k), []byte(vv)))
		_, ok := v.(string)
		if !ok {
			var builder strings.Builder
			builder.WriteString(`"`)
			builder.WriteString(k)
			builder.WriteString(`"`)
			k = builder.String()
		}
		r.Body = bytes.ReplaceAll(r.Body, []byte(k), []byte(vv))
		// 渲染头的value
		for headerK, headerV := range r.Header {
			r.Header[headerK] = string(bytes.ReplaceAll([]byte(headerV), []byte(k), []byte(vv)))
		}
	}
	for k, v := range parameters {
		if strings.HasPrefix(k, "{{FUNCTION.commander_sha256") {
			key := match(libs.ToStr(v))
			if key == "" {
				err = status.Errorf(codes.InvalidArgument, "commander_sha256 failed, accessKeySecret is nil")
				return err
			}
			requestBody := string(r.Body)
			nonce := gjson.Get(requestBody, "nonce")
			if !nonce.Exists() {
				err := status.Errorf(codes.InvalidArgument, "request body nonce is nil")
				return err
			}
			strNonce := libs.ToStr(nonce.Value())
			timestamp := gjson.Get(requestBody, "timestamp")
			if !timestamp.Exists() {
				err := status.Errorf(codes.InvalidArgument, "request body timestamp is nil")
				return err
			}
			strTimestamp := libs.ToStr(timestamp.Value())
			accessKeyId := gjson.Get(requestBody, "accessKeyId")
			if !accessKeyId.Exists() {
				err := status.Errorf(codes.InvalidArgument, "request body accessKeyId is nil")
				return err
			}
			message := []byte(key + strTimestamp + strNonce)
			h := hmac.New(sha256.New, []byte(key))
			h.Write(message)
			s := hex.EncodeToString(h.Sum(nil))
			r.Body = bytes.ReplaceAll(r.Body, []byte(k), []byte(fmt.Sprintf("%v", s)))
		}

	}
	return
}
