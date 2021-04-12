/**
* @Author: TongTongLiu
* @Date: 2019/11/27 5:17 下午
**/

package dao

import (
	"context"
	"fmt"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/siber"
	"testing"
)

func TestDao_SelectCaseLogs(t *testing.T) {
	c := &siber.CaseLog{
		//PlanId: "5dd26b5f9228c1001f2178f0",
		//FlowId: "5dd26b309228c1001f2178ef",
		CaseId: "5dd26b1c9228c1001f2178ee",
	}

	log, err := NewDao().ListCaseLog(context.Background(), c)
	fmt.Println(log)
	fmt.Println(err)
}
