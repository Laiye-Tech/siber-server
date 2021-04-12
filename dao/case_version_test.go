/**
* @Author: TongTongLiu
* @Date: 2019/12/11 12:23 下午
**/

package dao

import (
	"context"
	"reflect"
	"testing"
)

// 该function 用于将version信息从collection_case 中迁移至 collection_case_version 中
//func Test_MigrateCaseVersionInfo(t *testing.T) {
//	filter := &siber.FilterInfo{}
//	caseList, _, err := NewDao().ListCase(context.Background(), filter)
//	if err != nil {
//		fmt.Println("list case failed : ", err)
//		return
//	}
//	for _, c := range *caseList {
//		cInput := &CaseInfoStandard{
//			CaseId: c.CaseId,
//		}
//		cOutput, err := NewDao().SelectCase(context.Background(), cInput)
//		if err != nil {
//			fmt.Println("select case failed : ", cInput.CaseId, "err: ", err)
//			return
//		}
//		cVersion := &CaseVersionStandard{
//			CaseId:         cOutput.CaseId,
//			VersionControl: 3.0,
//			RequestHeader:  cOutput.RequestHeader,
//			RequestBody:    cOutput.RequestBody,
//			CheckPoint:     cOutput.CheckPoint,
//			InjectPoint:    cOutput.InjectPoint,
//			SleepPoint:     cOutput.SleepPoint,
//			Remark:         cOutput.Remark,
//			InsertTime:     cOutput.InsertTime,
//			UpdateTime:     cOutput.UpdateTime,
//		}
//		_, err = NewDao().InsertCaseVersion(context.Background(), cVersion)
//		if err != nil {
//			fmt.Println("insert case version failed : ", cInput.CaseId, "err: ", err)
//			return
//		}
//	}
//
//}

func TestDao_InsertCaseVersion(t *testing.T) {
	type args struct {
		ctx              context.Context
		caseVersionInput *CaseVersionStandard
	}
	tests := []struct {
		name                  string
		args                  args
		wantCaseVersionOutput *CaseVersionStandard
		wantErr               bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := &Dao{}
			gotCaseVersionOutput, err := dao.InsertCaseVersion(tt.args.ctx, tt.args.caseVersionInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertCaseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCaseVersionOutput, tt.wantCaseVersionOutput) {
				t.Errorf("InsertCaseVersion() gotCaseVersionOutput = %v, want %v", gotCaseVersionOutput, tt.wantCaseVersionOutput)
			}
		})
	}
}

func TestDao_SelectCaseVersion(t *testing.T) {
	type args struct {
		ctx              context.Context
		caseVersionInput *CaseVersionStandard
	}
	tests := []struct {
		name                  string
		args                  args
		wantCaseVersionOutput *CaseVersionStandard
		wantErr               bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dao := &Dao{}
			gotCaseVersionOutput, err := dao.SelectCaseVersion(tt.args.ctx, tt.args.caseVersionInput)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectCaseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCaseVersionOutput, tt.wantCaseVersionOutput) {
				t.Errorf("SelectCaseVersion() gotCaseVersionOutput = %v, want %v", gotCaseVersionOutput, tt.wantCaseVersionOutput)
			}
		})
	}
}
