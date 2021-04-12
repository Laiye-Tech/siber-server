package payload

import (
	"api-test/libs"
	"fmt"
	"github.com/bxcodec/faker/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var parameterRegex = regexp.MustCompile(fmt.Sprintf(`\{\{(%s)\.(.*?)\}\}`,
	strings.Join([]string{parameterTypeFunction, parameterTypeVariable}, "|")))

//var variablePlaceholderRegex = regexp.MustCompile(`(.*)\.(.*)`)
var variablePlaceholderRegex = regexp.MustCompile(`(.*)`)

type parameter interface {
	placeHolder() string
	Generate(ctx context.Context) (interface{}, error)
	Type() parameterType
	String() string
}

type parameterType string

const (
	parameterTypeFunction = "FUNCTION"
	parameterTypeVariable = "VARIABLE"
)

func newParameter(value string, typeStr string, placeholder string, variable *Variable) (parameter, error) {
	switch typeStr {
	case parameterTypeFunction:
		return newFunctionParameter(value, placeholder)
	case parameterTypeVariable:
		return newVariableParameter(placeholder, variable)
	}
	return nil, status.Errorf(codes.InvalidArgument, "invalid parameter type: %v", typeStr)
}

type variableParameter struct {
	parameter
	placeholder  string
	caseHashId   string
	variableName string
	variable     *Variable
}

func newVariableParameter(placeholder string, variable *Variable) (*variableParameter, error) {
	result := variablePlaceholderRegex.FindStringSubmatch(placeholder)
	if len(result) != 2 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid variable placeholder: %v", placeholder)
	}
	variableName := result[1]
	return &variableParameter{
		placeholder:  placeholder,
		variable:     variable,
		variableName: variableName}, nil
}

func (v *variableParameter) Type() parameterType {
	return parameterTypeVariable
}

func (v *variableParameter) String() string {
	return fmt.Sprintf("{{%s.%s}}", v.Type(), v.placeholder)
}

func (v *variableParameter) Generate(ctx context.Context) (interface{}, error) {
	value, err := v.variable.Get(ctx, v.variableName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "variable %v not found", v.placeholder)
	}
	return value, nil
}

type functionParameter struct {
	parameter
	value       string
	placeholder string
}

func newFunctionParameter(value string, placeholder string) (*functionParameter, error) {
	return &functionParameter{value: value, placeholder: placeholder}, nil
}

func (f *functionParameter) Type() parameterType {
	return parameterTypeFunction
}

func (f *functionParameter) String() string {
	return fmt.Sprintf("{{%s.%s}}", f.Type(), f.placeholder)
}
func match(s string) string {
	i := strings.Index(s, "(")
	if i >= 0 {
		j := strings.Index(s[i:], ")")
		if j >= 0 {
			return s[i+1 : j+i]
		}
	}
	return ""
}
func (f *functionParameter) Generate(ctx context.Context) (interface{}, error) {
	function := strings.Split(f.placeholder, "(")
	switch function[0] {
	case "random_string":
		value := match(f.placeholder)
		if value != "" {
			num, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "function err value %v", f.placeholder)
			}
			return libs.NewRandom().GetRandomString(num), nil
		} else {
			return libs.NewRandom().GetRandomString(16), nil
		}
	case "base_64":
		urlPath := match(f.placeholder)
		if urlPath != "" {
			base64Value, err := libs.RemoteBase64(urlPath)
			if err != nil {
				return nil, err
			}
			return base64Value, nil
		}
	case "random_phone":
		return libs.NewRandom().GetRandomPhone(), nil
	case "random_uuid":
		return faker.UUIDDigit(), nil
	case "random_email":
		return faker.Email(), nil
	case "random_url":
		return faker.URL(), nil
	case "random_name":
		return faker.FirstName(), nil
	case "unix_second":
		value := match(f.placeholder)
		if value == "str" {
			return libs.ToStr(time.Now().Unix()), nil
		}
		return time.Now().Unix(), nil
	case "unix_millisecond":
		value := match(f.placeholder)
		if value == "str" {
			return libs.ToStr(time.Now().UnixNano() / 1e6), nil
		}
		return time.Now().UnixNano() / 1e6, nil
	case "unix_microsecond":
		value := match(f.placeholder)
		if value == "str" {
			return libs.ToStr(time.Now().UnixNano() / 1e3), nil
		}
		return time.Now().UnixNano() / 1e3, nil
	case "unix_nanosecond":
		value := match(f.placeholder)
		if value == "str" {
			return libs.ToStr(time.Now().UnixNano()), nil
		}
		return time.Now().UnixNano(), nil
	case "unix_gmt_second":
		value := match(f.placeholder)
		if value == "str" {
			return libs.ToStr(time.Now().Unix() - 28800), nil
		}
		return time.Now().Unix() - 28800, nil
	}
	return f.value, nil
}

//func ExtractParameter(str string,variable *Variable) (parameters []parameter, err error) {
func ExtractParameter(ctx context.Context, str string, variable *Variable) (parameters map[string]interface{}, err error) {
	if str == "" {
		return
	}
	var parameterList []parameter
	result := parameterRegex.FindAllStringSubmatch(str, -1)
	for _, result := range result {
		if len(result) != 3 {
			return nil, status.Errorf(codes.InvalidArgument, "extract parameter error: %v", result)
		}
		p, err := newParameter(result[0], result[1], result[2], variable)
		if err != nil {
			return nil, err
		}
		parameterList = append(parameterList, p)
	}
	parameters = make(map[string]interface{})
	for _, p := range parameterList {
		parameters[p.String()], err = p.Generate(ctx)
		if err != nil {
			return
		}
	}
	return parameters, nil
}
