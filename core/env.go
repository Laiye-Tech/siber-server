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
func ManageEnvInfo(ctx context.Context, envInput *siber.ManageEnvInfo) (envOutput *siber.EnvInfo, err error) {
	// TODO: env 格式合理性检查
	if envInput == nil {
		return
	}
	switch envInput.ManageMode {
	case api.CreateItemMode:
		envOutput, err = dao.NewDao().InsertEnv(ctx, envInput.EnvInfo)
	case api.UpdateItemMode:
		envOutput, err = dao.NewDao().UpdateEnv(ctx, envInput.EnvInfo)
	case api.QueryItemMode:
		envOutput, err = dao.NewDao().SelectEnv(ctx, envInput.EnvInfo)
	case api.DeleteItemMode:
		envOutput, err = dao.NewDao().DeleteEnv(ctx, envInput.EnvInfo)
	}
	return envOutput, err
}

func ManageEnvList(ctx context.Context, request *siber.FilterInfo) (response *siber.EnvList, err error) {
	envList, totalNum, err := dao.NewDao().ListEnv(ctx, request)
	if err != nil {
		return
	}
	response = &siber.EnvList{
		EnvList:  *envList,
		TotalNum: uint32(totalNum),
	}
	return
}
