package payload

import (
	"fmt"
	"testing"
)

//func TestRequest_extractParameter(t *testing.T) {
//	tests := []struct {
//		request *Request
//		result  []string
//	}{
//		{
//			request: &Request{
//				Payload: Payload{Body: []byte(`{
//	"parameter1": "{{FUNCTION.random_string()}}",
//	"parameter2": "{{VARIABLE.case_hash.key1}}"
//}`)},
//			},
//			result: []string{"{{FUNCTION.random_string()}}", "{{VARIABLE.case_hash.key1}}"},
//		},
//	}
//	for _, test := range tests {
//		parameters, err := test.request.extractParameter(nil)
//		assert.Nil(t, err)
//		for i, parameter := range parameters {
//			assert.Equal(t, test.result[i], parameter.String())
//		}
//	}
//}
//
//func TestRequest_extractParameter_complex(t *testing.T) {
//	tests := []struct {
//		request *Request
//		result  []string
//	}{
//		{
//			request: &Request{
//				Payload: Payload{Body: []byte(`{
//	"parameter1": "{{{{FUNCTION.random_string()}}}}",
//	"parameter2": "{{VARIABLE.case_hash.key1}}"
//}`)},
//			},
//			result: []string{"{{FUNCTION.random_string()}}", "{{VARIABLE.case_hash.key1}}"},
//		},
//	}
//	for _, test := range tests {
//		parameters, err := test.request.extractParameter(nil)
//assert.Nil(t, err)
//		for i, parameter := range parameters {
//			assert.Equal(t, test.result[i], parameter.String())
//		}
//	}
//}

//func TestRequest_Render(t *testing.T) {
//	ctx := context.Background()
//	request := &Request{
//		Payload: Payload{Body: []byte(`{
//	"parameter1": "{{FUNCTION.random_string()}}",
//	"parameter2": "{{VARIABLE.case_hash.key1}}",
//	"parameter3": {{VARIABLE.case_hash.key2}}
//}`)},
//	}
//	v := &Variable{
//		values: map[string]map[string]interface{}{
//			"case_hash": {
//				"key1": "value1",
//				"key2": 1,
//			},
//		},
//	}
//	assert.Nil(t, request.Render(ctx, v))
//	m := map[string]interface{}{}
//	fmt.Println(string(request.Body))
//	assert.Nil(t, json.Unmarshal(request.Body, &m))
//	assert.Equal(t, "value1", m["parameter2"])
//	assert.Equal(t, 1, int(m["parameter3"].(float64)))
//}

func Test_printArray(t *testing.T) {
	list := []string{"tina", "liu"}
	str := fmt.Sprintf("%v", list)
	fmt.Println(str)
}
