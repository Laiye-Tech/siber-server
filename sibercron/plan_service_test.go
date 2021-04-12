/**
* @Author: TongTongLiu
* @Date: 2020/5/15 7:50 下午
**/

package sibercron

import (
	"api-test/initial"
	"context"
	"testing"
)

func TestRefreshAllPlanServices(t *testing.T) {
	initial.Initial()
	SetServiceRoutes(context.Background())
	RefreshAllPlanServices(context.Background())
}

