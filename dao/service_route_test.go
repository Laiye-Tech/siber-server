/**
* @Author: TongTongLiu
* @Date: 2020/5/15 4:00 下午
**/

package dao

import (
	"api-test/initial"
	"context"
	"fmt"
	"testing"
)

func TestListServiceRoute(t *testing.T) {
	initial.Initial()
	services, err := NewDao().ListServiceRoute(context.Background())
	fmt.Println(err)
	fmt.Println(services)
}
