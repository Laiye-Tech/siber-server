package core

import (
	"api-test/configs"
	"api-test/libs"
	"context"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"github.com/patrickmn/go-cache"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type ScheduleManager struct {
	cron  *cron.Cron
	cache *cache.Cache
}

func NewScheduleManager() *ScheduleManager {
	instance := &ScheduleManager{}
	instance.cron = cron.New(cron.WithSeconds())
	instance.cache = cache.New(cache.NoExpiration, 10*time.Minute)
	return instance
}

func cornRunPlan(ctx context.Context, runPlanInfo *siber.RunPlanRequest) {
	p, _ := DescribePlan(ctx, runPlanInfo, CronTrigger)
	_, _ = p.Run(ctx)
}

// 开启 cron
func (m *ScheduleManager) Run(ctx context.Context) error {
	if configs.AppDebug() {
		return nil
	}
	// 项目启起来就刷新cron
	_ = m.cronRun(ctx)
	m.cron.Run()

	// 定时刷新cron
	_, err := m.cron.AddFunc("13 13 * * * ?", func() {
		_ = m.cronRun(ctx)
	})
	if err != nil {
		xzap.Logger(ctx).Info("init refresh cron failed", zap.Any("err", err))
	}
	return err
}

// 根据持久化的plan list 进行cron list 的订正
func (m *ScheduleManager) cronRun(ctx context.Context) (err error) {
	// 测试环境不开启cron
	if configs.AppDebug() {
		return
	}
	var filter = map[string]string{"action": "cron", "page": "1", "page_size": "10000"}
	res, err := ManagePlanList(ctx, &siber.FilterInfo{
		FilterContent: filter,
	})
	if err != nil {
		return err
	}
	PlanList := make(map[string]string)
	for _, v := range res.PlanInfoList {
		for _, t := range v.TriggerCondition {
			PlanList[t.EnvironmentName+v.PlanId] = t.TriggerCron
		}
		_ = m.EditCron(ctx, v)
	}
	Items := m.cache.Items()
	for k := range Items {
		if _, ok := PlanList[k]; !ok {
			err = m.deleteCron(ctx, k)
			if err != nil {
				xzap.Logger(ctx).Error("revision cron task error", zap.Any("err:", err))
			}
		}
	}
	return
}

func (m *ScheduleManager) EditCron(ctx context.Context, planInfo *siber.PlanInfo) (err error) {
	if configs.AppDebug() {
		return
	}
	m.deleteAllCron(ctx, planInfo)
	for _, v := range planInfo.TriggerCondition {
		// TODO: 对 cond 格式做校验，格式错误，报错。正确格式："" 或者标准的cron 写法
		envPlanId := v.EnvironmentName + planInfo.PlanId
		_, found := m.cache.Get(envPlanId)
		if found {
			if v.TriggerCron != "" {
				err = m.updateCron(ctx, envPlanId, v.TriggerCron)
				if err == nil {
					continue
				}
				err = status.Errorf(codes.InvalidArgument, "add cron task error %v: ", err)
				xzap.Logger(ctx).Error("add cron task failed",
					zap.Any("envPlanId", envPlanId),
					zap.Any("err", err))
				return
			}
			if v.TriggerCron == "" {
				err = m.deleteCron(ctx, envPlanId)
			}
		} else {
			if v.TriggerCron == "" {
				continue
			}
			xzap.Logger(ctx).Info("the add cron spec ", zap.String(envPlanId, v.TriggerCron))
			id, err := m.cron.AddFunc(v.TriggerCron, func() {
				cornRunPlan(ctx, &siber.RunPlanRequest{
					PlanInfo:         planInfo,
					TriggerCondition: v,
				})
			})
			if err != nil {
				err = status.Errorf(codes.InvalidArgument, "add cron task error %v: ", err)
				xzap.Logger(ctx).Error("add cron task failed",
					zap.String("envPlanId", envPlanId),
					zap.Any("err", err))
				return err
			}
			m.cache.Set(envPlanId, id, 0)
		}

	}
	return
}

//
func (m *ScheduleManager) updateCron(ctx context.Context, envPlanId string, spec string) (err error) {
	id, found := m.cache.Get(envPlanId)
	if !found {
		return
	}
	err = m.cron.UpdateScheduleWithSpec(id.(cron.EntryID), spec)
	xzap.Logger(ctx).Info("the cron update spec ", zap.String(envPlanId, spec))
	if err != nil {
		err = status.Errorf(codes.Unknown, "update cron task error %v: ", err)
		xzap.Logger(ctx).Error("update cron task error %v: ", zap.Any("err", err))
		return
	}
	return

}

// 从cron list 删除cron，修改plan，或者定时修正时触发
func (m *ScheduleManager) deleteCron(ctx context.Context, envPlanId string) (err error) {
	if configs.AppDebug() {
		return
	}
	if envPlanId == "" {
		return
	}

	id, found := m.cache.Get(envPlanId)
	if found {
		xzap.Logger(ctx).Info("the cron delete spec ", zap.String("envPlanId", envPlanId))
		m.cron.Remove(id.(cron.EntryID))
		m.cache.Delete(envPlanId)
		return
	}
	return

}

//根据现有plan env配置取交集进行删除
func (m *ScheduleManager) deleteAllCron(ctx context.Context, planInfo *siber.PlanInfo) {
	if configs.AppDebug() {
		return
	}
	if planInfo.PlanId == "" {
		return
	}
	envList := []string{"dev", "test", "stage", "prod"}
	PlanEnvList := make([]string, 0)
	if len(planInfo.TriggerCondition) > 0 {
		for _, v := range planInfo.TriggerCondition {
			PlanEnvList = append(PlanEnvList, v.EnvironmentName)
		}
	}
	envList = libs.Difference(envList, PlanEnvList)
	for _, v := range envList {
		envPlanId := v + planInfo.PlanId
		id, found := m.cache.Get(envPlanId)
		if found {
			xzap.Logger(ctx).Info("the cron delete spec ", zap.String("envPlanId", envPlanId))
			m.cron.Remove(id.(cron.EntryID))
			m.cache.Delete(envPlanId)
		}
	}
	return
}
