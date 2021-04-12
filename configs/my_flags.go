/**
* @Author: TongTongLiu
* @Date: 2019-08-07 12:27
**/

package configs

import (
	"flag"
)

var (
	//FlagLogPidName log文件名
	FlagLogPidName string
	// FlagConfigPath 配置文件路径
	FlagConfigPath string
	// FlagPort 主服务端口
	FlagPort int
	// FlagGWPort grpc-gateway端口
	FlagGWPort int
)

//FlagParse 解析命令行参数
func FlagParse() {
	if !flag.Parsed() {
		lName := flag.String("log", "", "the different process log suffix")
		configPath := flag.String("conf", "", "the path of config file")
		grpcPort := flag.Int("port", 0, "main server port")
		gwPort := flag.Int("gwport", 0, "grpc gateway server port")
		flag.Parse()
		FlagLogPidName = *lName
		FlagConfigPath = *configPath
		FlagPort = *grpcPort
		FlagGWPort = *gwPort
	}

}
