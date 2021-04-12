/**
* @Author: TongTongLiu
* @Date: 2019/12/11 7:22 下午
**/

package core

import (
	"api-test/api"
	"api-test/dao"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sort"
	"time"
)

type VersionDetail struct {
	CurrentVersion string
	AfterVersion   int32
}

// 对已有的case version 进行排序
func sortCaseVersion(ctx context.Context, c *Case) (versionInfos []VersionDetail, err error) {
	if c == nil || c.Id == "" {
		return
	}
	caseVersionInput := &dao.CaseVersionStandard{
		CaseId: c.Id,
	}
	versionListTmp, err := dao.NewDao().ListCaseVersionByID(ctx, caseVersionInput)
	if err != nil {
		return
	}
	if versionListTmp == nil || len(*versionListTmp) == 0 {
		xzap.Logger(ctx).Warn("unMatch case version")
		err = status.Errorf(codes.OutOfRange, "unMatch case version,caseId:%s", c.Id)
		return
	}
	for _, v := range *versionListTmp {
		afterVersion, err := api.RegexVersion(v.VersionControl)
		if err != nil {
			return nil, err
		}
		versionInfos = append(versionInfos, VersionDetail{CurrentVersion: v.VersionControl, AfterVersion: afterVersion})
	}
	sort.Slice(versionInfos, func(i, j int) bool {
		return versionInfos[i].AfterVersion > versionInfos[j].AfterVersion
	})
	if versionInfos == nil {
		xzap.Logger(ctx).Warn("versionInfos is nil")
		err = status.Errorf(codes.OutOfRange, "versionInfos is nil")
		return
	}
	return
}

// 维护case版本内信息，如request，checkpoint等
func ManageCaseVersion(ctx context.Context, caseVersionInput *siber.ManageCaseVersionInfo) (caseVersionOutput *siber.CaseVersionInfo, err error) {
	// TODO: case 格式合理性检查：1-循环依赖 2-依赖case是否存在指定变量
	if caseVersionInput == nil || caseVersionInput.CaseVersion == nil {
		err = status.Errorf(codes.InvalidArgument, "ManageCaseVersion failed, caseVersion is nil")
		return
	}
	caseInputStandard, err := CaseVersionToStandard(ctx, caseVersionInput.CaseVersion)
	if err != nil {
		return
	}
	caseOutputStandard := new(dao.CaseVersionStandard)
	switch caseVersionInput.ManageMode {
	// update 实现有则更新，无则插入的功能
	case api.UpdateItemMode:
		caseOutputStandard, err = dao.NewDao().UpdateCaseVersion(ctx, caseInputStandard)
		if err != nil || caseOutputStandard == nil {
			return
		}
		// 更新collection_case 的最后更新时间，不做强制要求
		caseInput := &siber.CaseInfo{
			CaseId:     caseOutputStandard.CaseId,
			UpdateTime: time.Now().Unix(),
		}
		_ = dao.NewDao().UpdateCaseTime(ctx, caseInput)
	case api.QueryItemMode:
		caseOutputStandard, err = dao.NewDao().SelectCaseVersion(ctx, caseInputStandard)
	case api.DeleteItemMode:
		caseOutputStandard, err = dao.NewDao().DeleteCaseVersion(ctx, caseInputStandard)
	}
	if err != nil {
		return
	}
	// TODO: 将 standard 的 output 转化为proto可以识别的
	caseVersionOutput, err = CaseVersionToProto(ctx, caseOutputStandard)
	return
}
