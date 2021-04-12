package payload

import (
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// map[case_hash_id]map[name]interface{}
type Variable struct {
	values map[string]interface{}
}

func (v *Variable) Create() {
	v.values = make(map[string]interface{})
	return
}

func (v *Variable) Get(ctx context.Context, name string) (interface{}, error) {
	if _, ok := v.values[name]; !ok {
		xzap.Logger(ctx).Info("variable not defined", zap.Any("name", name))
		return nil, nil
	}
	value := v.values[name]
	xzap.Logger(ctx).Info("get variable success", zap.Any("name", name), zap.Any("value:", value))

	return value, nil
}

func (v *Variable) Set(ctx context.Context, name string, value interface{}) {
	v.values[name] = value
	xzap.Logger(ctx).Info("set variable success", zap.Any("name", name), zap.Any("value:", value))

	return
}
