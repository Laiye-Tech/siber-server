package payload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const unsupportedError = "The value %v and return value %v type is unsupported"
const includeError = "The value %v is not included in the return value %v"
const (
	GreaterThan          = ">"
	LessThan             = "<"
	GreaterThanOrEqualTo = ">="
	LessThanOrEqualTo    = "<="
)

type Checker interface {
	Render(ctx context.Context, v *Variable) (err error)
	Create()
	GetCheckerName() string
	Check(keyExists bool, value interface{}) error
}

type ExistChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue bool
}

type LengthChecker struct {
	Checker `json:"Checker,omitempty"`
	Name    string
	Length  float64
}

type RegexChecker struct {
	Checker `json:"Checker,omitempty"`
	Name    string
	regex   *regexp.Regexp
}

type EqualChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue interface{}
}

type NotEqualChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue interface{}
}

type InChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue interface{}
}

type IncludeChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue interface{}
}

type NotIncludeChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue interface{}
}

type ComparisonChecker struct {
	Checker     `json:"Checker,omitempty"`
	Name        string
	ExpectValue interface{}
}

func checkParameter(k string) (ok bool) {
	if (strings.HasPrefix(k, "{{VARIABLE")) || (strings.HasPrefix(k, "{{VARIABLE")) && strings.HasSuffix(k, "}}") {
		return true
	}
	return false
}

func renderParameter(ctx context.Context, expectValue interface{}, v *Variable) (renderExpectValue interface{}, err error) {
	// 渲染expectValue
	str, ok := (expectValue).(string)
	if !ok {
		renderExpectValue = expectValue
		return renderExpectValue, err
	}
	parameters, err := ExtractParameter(ctx, str, v)
	if err != nil {
		return
	}
	if len(parameters) == 0 {
		renderExpectValue = expectValue
	}
	for k, v := range parameters {
		if checkParameter(k) {
			renderExpectValue = v
		} else {
			expectByte := bytes.ReplaceAll([]byte(str), []byte(k), []byte(fmt.Sprintf("%v", v)))
			renderExpectValue = string(expectByte)
		}
	}
	return renderExpectValue, err
}

func contains(values []interface{}, ExceptValue interface{}) bool {
	for _, value := range values {
		if value == ExceptValue {
			return true
		}
	}
	return false
}

func comparisonValue(value float64, ExceptValue float64, symbol string) (err error) {
	if symbol == GreaterThan {
		if value > ExceptValue {
			return
		}
		return status.Errorf(codes.InvalidArgument, "The value %v not greater than return value %v", value, ExceptValue)
	}
	if symbol == GreaterThanOrEqualTo {
		if value >= ExceptValue {
			return
		}
		return status.Errorf(codes.InvalidArgument, "The value %v not greater than or equal to return value %v", value, ExceptValue)
	}
	if symbol == LessThan {
		if value < ExceptValue {
			return
		}
		return status.Errorf(codes.InvalidArgument, "The value %v not less than return value %v", value, ExceptValue)
	}
	if symbol == LessThanOrEqualTo {
		if value <= ExceptValue {
			return
		}
		return status.Errorf(codes.InvalidArgument, "The value %v not less than or equal to return value %v", value, ExceptValue)
	}
	return status.Errorf(codes.InvalidArgument, unsupportedError, ExceptValue, value)
}

func (comparisonChecker *ComparisonChecker) Render(ctx context.Context, v *Variable) (err error) {
	// 大于渲染
	//
	renderExpectValue, err := renderParameter(ctx, comparisonChecker.ExpectValue, v)
	if err != nil {
		return
	}
	comparisonChecker.ExpectValue = renderExpectValue
	return
}
func (comparisonChecker *ComparisonChecker) GetCheckerName() string {
	return "ComparisonChecker"
}

func (equalChecker *EqualChecker) Render(ctx context.Context, v *Variable) (err error) {
	// 等于渲染
	renderExpectValue, err := renderParameter(ctx, equalChecker.ExpectValue, v)
	if err != nil {
		return
	}
	equalChecker.ExpectValue = renderExpectValue
	return
}
func (equalChecker *EqualChecker) GetCheckerName() string {
	return "EqualChecker"
}

func (notEqualChecker *NotEqualChecker) Render(ctx context.Context, v *Variable) (err error) {
	renderExpectValue, err := renderParameter(ctx, notEqualChecker.ExpectValue, v)
	if err != nil {
		return
	}
	notEqualChecker.ExpectValue = renderExpectValue
	return
}

func (notEqualChecker *NotEqualChecker) GetCheckerName() string {
	return "NotEqualChecker"
}

func (includeChecker *IncludeChecker) Render(ctx context.Context, v *Variable) (err error) {
	// 包含渲染
	renderExpectValue, err := renderParameter(ctx, includeChecker.ExpectValue, v)
	if err != nil {
		return
	}
	includeChecker.ExpectValue = renderExpectValue
	return
}

func (includeChecker *IncludeChecker) GetCheckerName() string {
	return "IncludeChecker"
}

func (notIncludeChecker *NotIncludeChecker) Render(ctx context.Context, v *Variable) (err error) {
	// 不包含渲染
	renderExpectValue, err := renderParameter(ctx, notIncludeChecker.ExpectValue, v)
	if err != nil {
		return
	}
	notIncludeChecker.ExpectValue = renderExpectValue
	return
}

func (notIncludeChecker *NotIncludeChecker) GetCheckerName() string {
	return "NotIncludeChecker"
}

func (inChecker *InChecker) Render(ctx context.Context, v *Variable) (err error) {
	// 在....里 渲染 ,
	renderExpectValue, err := renderParameter(ctx, inChecker.ExpectValue, v)
	if err != nil {
		return
	}
	inChecker.ExpectValue = renderExpectValue
	return
}
func (inChecker *InChecker) GetCheckerName() string {
	return "InChecker"
}

func (existChecker *ExistChecker) Render(ctx context.Context, v *Variable) (err error) {
	return
}
func (existChecker *ExistChecker) GetCheckerName() string {
	return "ExistChecker"
}

func (lengthChecker *LengthChecker) Render(ctx context.Context, v *Variable) (err error) {
	return
}

func (lengthChecker *LengthChecker) GetCheckerName() string {
	return "LengthChecker"
}

func (comparisonChecker ComparisonChecker) Check(keyExists bool, value interface{}) (err error) {

	valFloat, ok := value.(int)
	if ok {
		value = float64(valFloat)
	}
	switch value.(type) {
	case float64:
		value, ok := value.(float64)
		ExpectValue, okk := comparisonChecker.ExpectValue.(float64)
		if !ok && !okk {
			return status.Errorf(codes.InvalidArgument, unsupportedError, comparisonChecker.ExpectValue, value)
		}
		err = comparisonValue(value, ExpectValue, comparisonChecker.Name)
		return
	}
	return status.Errorf(codes.InvalidArgument, unsupportedError, comparisonChecker.ExpectValue, value)
}

func (lengthChecker LengthChecker) Check(keyExists bool, value interface{}) (err error) {
	length := -1
	strValue, ok := value.(string)
	if ok {
		length = strings.Count(strValue, "") - 1
	}

	intValue, ok := value.(int)
	if ok {
		length = strings.Count(strconv.Itoa(intValue), "") - 1
	}

	boolValue, ok := value.(bool)
	if ok {
		length = strings.Count(strconv.FormatBool(boolValue), "") - 1
	}

	listValue, ok := value.([]interface{})
	if ok {
		length = len(listValue)
	}

	if length == -1 {
		return status.Errorf(codes.InvalidArgument, "unsupported type for length :%s", reflect.TypeOf(value))
	}
	if lengthChecker.Length == float64(length) {
		return nil
	} else {
		return status.Errorf(codes.InvalidArgument, "The length %v does not match the return value length %v", lengthChecker.Length, value)
	}
}

// 判断 value 是否在 inChecker的InValue中
func (inChecker InChecker) Check(keyExists bool, value interface{}) (err error) {
	// 支持判断是否在数组中

	listValue, ok := inChecker.ExpectValue.([]interface{})
	if ok {
		for _, v := range listValue {
			if v == value {
				return
			}
		}
		return status.Errorf(codes.InvalidArgument, "value not in the list")
	}
	// 判断是否包含子字符串
	strValue, ok := value.(string)
	inValue, okk := inChecker.ExpectValue.(string)
	if ok && okk {
		// 如果是可以转换为json，要以json类型做比对
		var jsonValue map[string]interface{}
		var jsonExpect map[string]interface{}
		err = json.Unmarshal([]byte(strValue), &jsonValue)
		err2 := json.Unmarshal([]byte(inValue), &jsonExpect)
		if err == nil && err2 == nil {
			for k := range jsonValue {
				if _, ok := jsonExpect[k]; !ok {
					err = status.Errorf(codes.FailedPrecondition, "key %s not exists", k)
					return
				}
				// 为防止类型不一样，这里不用 "=" 做判断
				if diff := cmp.Diff(jsonValue[k], jsonExpect[k]); diff != "" {
					return status.Errorf(codes.InvalidArgument, "key %s value not equal", k, diff)
				}
			}
			return
		}
		// 忽略转换json报的错
		err = nil

		if strings.ContainsAny(inValue, strValue) {
			return
		}
		return status.Errorf(codes.InvalidArgument, "string not contained, %s %s", inValue, value)
	}

	return status.Errorf(codes.InvalidArgument, "unsupported in check type")
}

// 判断value（返回值）是否包含includeChecker的ExpectValue
func (includeChecker IncludeChecker) Check(keyExists bool, value interface{}) (err error) {
	//value 接口返回值
	//ExpectValue 校验值
	switch value.(type) {
	case string:
		strExpectValue, ok := includeChecker.ExpectValue.(string)
		if ok {
			//如果二者都能够转成json类型，则以json类型做比对
			var jsonValue map[string]interface{}
			var jsonExpect map[string]interface{}
			strValue, ok := value.(string)
			// 理论上ok是不会为false的
			if !ok {
				return
			}
			err = json.Unmarshal([]byte(strValue), &jsonValue)
			err2 := json.Unmarshal([]byte(strExpectValue), &jsonExpect)
			if err == nil && err2 == nil {
				for k := range jsonExpect {
					if _, ok := jsonValue[k]; !ok {
						err = status.Errorf(codes.FailedPrecondition, "key %s not exists", k)
						return
					}
					// 为防止类型不一样，这里不用 "=" 做判断
					if diff := cmp.Diff(jsonValue[k], jsonExpect[k]); diff != "" {
						return status.Errorf(codes.InvalidArgument, "key %s value not equal", k, diff)
					}
				}
				return
			}
			// 忽略转换json报的错
			err = nil

			ok = strings.ContainsAny(value.(string), strExpectValue)
			if !ok {
				return status.Errorf(codes.InvalidArgument, includeError, includeChecker.ExpectValue, value)
			}
			return nil
		}
		return status.Errorf(codes.InvalidArgument, unsupportedError, includeChecker.ExpectValue, value)
	case []interface{}:
		sliceValue, ok := value.([]interface{})
		if !ok {
			return status.Errorf(codes.InvalidArgument, unsupportedError, includeChecker.ExpectValue, value)
		}
		strExpectValue, ok := includeChecker.ExpectValue.(string)
		if ok {
			if contains(sliceValue, strExpectValue) {
				return
			}
			return status.Errorf(codes.InvalidArgument, includeError, includeChecker.ExpectValue, value)

		}
		intExpectValue, ok := includeChecker.ExpectValue.(float64)
		if ok {
			if contains(sliceValue, intExpectValue) {
				return
			}
			return status.Errorf(codes.InvalidArgument, includeError, includeChecker.ExpectValue, value)

		}
		SliceExpectValue, ok := includeChecker.ExpectValue.([]interface{})
		if ok {
			if len(SliceExpectValue) > len(sliceValue) {
				return status.Errorf(codes.InvalidArgument, "check value length cannot be greater than return value length")
			}
			for _, e := range SliceExpectValue {
				if !contains(sliceValue, e) {
					return status.Errorf(codes.InvalidArgument, includeError, includeChecker.ExpectValue, value)
				}
			}
			return
		}
		return status.Errorf(codes.InvalidArgument, unsupportedError, includeChecker.ExpectValue, value)
	}
	return status.Errorf(codes.InvalidArgument, unsupportedError, includeChecker.ExpectValue, value)
}
func (notIncludeChecker NotIncludeChecker) Check(keyExists bool, value interface{}) (err error) {
	//value 接口返回值
	//ExpectValue 校验值
	switch value.(type) {
	case string:
		strExpectValue, ok := notIncludeChecker.ExpectValue.(string)
		if ok {
			ok := strings.ContainsAny(value.(string), strExpectValue)
			if ok {
				return status.Errorf(codes.InvalidArgument, includeError, notIncludeChecker.ExpectValue, value)
			}
			return nil
		}
		return status.Errorf(codes.InvalidArgument, unsupportedError, notIncludeChecker.ExpectValue, value)

	case []interface{}:
		sliceValue, ok := value.([]interface{})
		if !ok {
			return status.Errorf(codes.InvalidArgument, unsupportedError, notIncludeChecker.ExpectValue, value)
		}
		strExpectValue, ok := notIncludeChecker.ExpectValue.(string)
		if ok {
			if !contains(sliceValue, strExpectValue) {
				return nil
			}
			return status.Errorf(codes.InvalidArgument, includeError, notIncludeChecker.ExpectValue, value)

		}
		intExpectValue, ok := notIncludeChecker.ExpectValue.(float64)
		if ok {
			if !contains(sliceValue, intExpectValue) {
				return nil
			}
			return status.Errorf(codes.InvalidArgument, includeError, notIncludeChecker.ExpectValue, value)
		}
		SliceExpectValue, ok := notIncludeChecker.ExpectValue.([]interface{})
		if ok {
			if len(SliceExpectValue) > len(sliceValue) {
				return status.Errorf(codes.InvalidArgument, "check value length cannot be greater than return value length")
			}
			for _, e := range SliceExpectValue {
				if contains(sliceValue, e) {
					return status.Errorf(codes.InvalidArgument, includeError, notIncludeChecker.ExpectValue, value)
				}
			}
			return
		}
		return status.Errorf(codes.InvalidArgument, unsupportedError, notIncludeChecker.ExpectValue, value)
	}
	return status.Errorf(codes.InvalidArgument, unsupportedError, notIncludeChecker.ExpectValue, value)
}

func (regexChecker RegexChecker) Check(keyExists bool, value interface{}) (err error) {
	strValue, ok := value.(string)
	if !ok {
		return errors.New("value is must be string")
	}
	matchBool := regexChecker.regex.MatchString(strValue)
	if matchBool == true {
		return nil
	} else {
		return status.Errorf(codes.InvalidArgument, "the regex %v does not match the return value %v", regexChecker.regex.String(), value)
	}
}

func (equalChecker EqualChecker) Check(keyExists bool, value interface{}) (err error) {
	// 所有的数值类型，转为int64判断。因为：structValue是用float64
	valFloat, ok := value.(int)
	if ok {
		value = float64(valFloat)
	}
	//如果能够转成json类型，则以json类型做比对
	strValue, ok := value.(string)
	strExpect, okk := equalChecker.ExpectValue.(string)
	if ok && okk {
		var jsonValue map[string]interface{}
		var jsonExpect map[string]interface{}
		err = json.Unmarshal([]byte(strValue), &jsonValue)
		err2 := json.Unmarshal([]byte(strExpect), &jsonExpect)
		if err == nil && err2 == nil {
			if diff := cmp.Diff(jsonValue, jsonExpect); diff != "" {
				return status.Errorf(codes.InvalidArgument, "The json value does not equal return value .Diff info: %v ", equalChecker.ExpectValue, value, diff)
			}
			return
		}
		// 忽略转换json报的错
		err = nil
	}

	// 其他类型比对：非json的str int list
	if diff := cmp.Diff(equalChecker.ExpectValue, value); diff != "" {
		return status.Errorf(codes.InvalidArgument, "The value %v does not equal return value %v. \n Diff info: %v ", equalChecker.ExpectValue, value, diff)
	}
	return
}

func (notEqualChecker NotEqualChecker) Check(keyExists bool, value interface{}) (err error) {
	// 所有的数值类型，转为int64判断。因为：structValue是用float64
	valFloat, ok := value.(int)
	if ok {
		value = float64(valFloat)
	}
	//如果能够转成json类型，则以json类型做比对
	strValue, ok := value.(string)
	strExpect, okk := notEqualChecker.ExpectValue.(string)
	if ok && okk {
		var jsonValue map[string]interface{}
		var jsonExpect map[string]interface{}
		err = json.Unmarshal([]byte(strValue), &jsonValue)
		err2 := json.Unmarshal([]byte(strExpect), &jsonExpect)
		if err == nil && err2 == nil {
			if diff := cmp.Diff(jsonValue, jsonExpect); diff == "" {
				return status.Errorf(codes.InvalidArgument, "The json value %v equal return value %v. \n Diff info: %v ", notEqualChecker.ExpectValue, value, diff)
			}
			return
		}
		// 忽略转换json报的错
		err = nil
	}

	// 其他类型比对：非json的str int list
	if diff := cmp.Diff(notEqualChecker.ExpectValue, value); diff == "" {
		return status.Errorf(codes.InvalidArgument, "The value %v equal return value %v. \n Diff info: %v ", notEqualChecker.ExpectValue, value, diff)
	}
	return
}

func (existChecker ExistChecker) Check(keyExists bool, value interface{}) (err error) {
	if !existChecker.ExpectValue && !keyExists {
		return
	}
	if existChecker.ExpectValue && keyExists {
		return
	}
	return status.Errorf(codes.InvalidArgument, "The key is not exist")

}
