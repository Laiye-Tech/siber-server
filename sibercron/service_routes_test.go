/**
* @Author: TongTongLiu
* @Date: 2020/5/15 4:17 下午
**/

package sibercron

import (
	"api-test/initial"
	"context"
	"fmt"
	"testing"
)

func TestSetServiceRoutes(t *testing.T) {
	initial.Initial()
	SetServiceRoutes(context.Background())
	s := Services
	fmt.Println(s)
}