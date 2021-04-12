package payload

import (
	regexp "regexp"
	"testing"
)

//func TestLengthChecker_Check(t *testing.T) {
//	type fields struct {
//		Checker Checker
//		length  int
//	}
//	type args struct {
//		value interface{}
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{"common value", fields{length: 3}, args{value: "123"}, false},
//		{"common value", fields{length: 3}, args{value: "aabbcc"}, true},
//		{"value is empty string", fields{length: 0}, args{value: ""}, false},
//		{"value type ", fields{length: 0}, args{value: 123}, true},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			lengthChecker := LengthChecker{
//				Checker: tt.fields.Checker,
//				length:  tt.fields.length,
//			}
//			if err := lengthChecker.Check(tt.args.value); (err != nil) != tt.wantErr {
//				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestRegexChecker_Check(t *testing.T) {
	type fields struct {
		Checker Checker
		regex   *regexp.Regexp
	}
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"error regex", fields{regex: regexp.MustCompile(`^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`)}, args{value: "10086"}, true},
		{"phone regex", fields{regex: regexp.MustCompile(`^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`)}, args{value: "13800138000"}, false},
		{"error value type", fields{regex: regexp.MustCompile(`[a-z]`)}, args{value: 123}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexChecker := RegexChecker{
				Checker: tt.fields.Checker,
				regex:   tt.fields.regex,
			}
			if err := regexChecker.Check(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func TestEqualChecker_Check(t *testing.T) {
//	type fields struct {
//		Checker     Checker
//		formatValue interface{}
//	}
//	type args struct {
//		value interface{}
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{"not equal", fields{formatValue: "1"}, args{value: "123"}, true},
//		{"str equal", fields{formatValue: "123"}, args{value: "123"}, false},
//		{"json equal", fields{formatValue: `{"data":"123"}`}, args{value: `{"data":"123"}`}, false},
//		{"int equal", fields{formatValue: 123}, args{value: 123}, false},
//		{"bool not equal", fields{formatValue: true}, args{value: true}, true},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			equalChecker := EqualChecker{
//				Checker:     tt.fields.Checker,
//				formatValue: tt.fields.formatValue,
//			}
//			if err := equalChecker.Check(tt.args.value); (err != nil) != tt.wantErr {
//				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestExistChecker_Check(t *testing.T) {
	type fields struct {
		Checker Checker
	}
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"exits", fields{}, args{value: ""}, false},
		{"exits", fields{}, args{value: "213"}, false},
		{"not exits", fields{}, args{value: nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existChecker := ExistChecker{
				Checker: tt.fields.Checker,
			}
			if err := existChecker.Check(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
