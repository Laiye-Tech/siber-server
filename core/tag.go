package core

import (
	"api-test/api"
	"api-test/dao"
	"context"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
)

/*
* 维护plan：CREATE：创建，UPDATE：修改
 */
func ManageTagInfo(ctx context.Context, tagInput *siber.ManageTagInfo) (tagOutput *siber.TagInfo, err error) {
	if tagInput == nil {
		return
	}
	switch tagInput.ManageMode {
	case api.CreateItemMode:
		tagOutput, err = dao.NewDao().InsertTag(ctx, tagInput.TagInfo)
	case api.UpdateItemMode:
		tagOutput, err = dao.NewDao().UpdateTag(ctx, tagInput.TagInfo)
	case api.QueryItemMode:
		tagOutput, err = dao.NewDao().SelectTag(ctx, tagInput.TagInfo)
	case api.DeleteItemMode:
		tagOutput, err = dao.NewDao().DeleteTag(ctx, tagInput.TagInfo)
	}
	return tagOutput, err
}

func ManageTagList(ctx context.Context, request *siber.FilterInfo) (response *siber.TagList, err error) {
	tagList, totalNum, err := dao.NewDao().ListTag(ctx, request)
	if err != nil {
		return
	}
	response = &siber.TagList{
		TagList:  *tagList,
		TotalNum: uint32(totalNum),
	}
	return
}
