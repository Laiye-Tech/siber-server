/**
* @Author: TongTongLiu
* @Date: 2019/12/31 11:51 上午
**/

package initial

import (
	"api-test/configs"
	"api-test/libs"
)

func Initial() {
	configs.InitConfigs()
	libs.InitLog()
	return
}
