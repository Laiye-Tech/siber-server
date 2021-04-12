/**
* @Author: TongTongLiu
* @Date: 2019-08-07 12:24
**/

package configs

import (
	"fmt"
	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"net"
	"os"

	"github.com/BurntSushi/toml"
)

var (
	globalConf         = Config{}
	testConfigsMapping = map[string]string{
		//liutongtong
		"64:5a:ed:eb:01:d9": "/Users/liutongtong/go/src/api-test/configs/local_ttl.toml",
		"6c:96:cf:dd:18:4d": "/Users/dongmengnan/Works/programs/SaaS/src/api-test/configs/mengnan.toml",
		"64:5a:ed:eb:19:b1": "/Users/jichen/Dropbox/Coding/api-test/configs/jichen.toml",
		"64:5a:ed:e9:e5:15": "/Users/yunqiang/Desktop/api-test/configs/local_ttl.toml",
	}
	debug = true
)

func GetAppName() string {
	return globalConf.AppName
}

func AppDebug() bool {
	return globalConf.Debug
}

func IsLocal() bool {
	return getMacAddr() == "6c:96:cf:dd:18:4d"
}

func getMacAddr() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("get mac addr er: " + err.Error())
	}
	macAddr := ""
	for _, inter := range interfaces {
		if inter.Name == "en0" {
			macAddr = inter.HardwareAddr.String()
		}
	}
	return macAddr
}

func GetInternalIp() string {
	var ip string
	ip = getEnvIp()
	if ip != "" {
		return ip
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	} else {
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}
	}
	return ip
}

func getEnvIp() string {
	return os.Getenv("HOST_IP")
}

func InitConfigs() {
	FlagParse()
	var configFile string
	if FlagConfigPath == "" && debug {
		configFile = testConfigsMapping[getMacAddr()]
	} else {
		configFile = FlagConfigPath
	}
	_, err := os.Stat(configFile)
	if err != nil && os.IsNotExist(err) {
		panic(fmt.Sprintf("config file (%s) not exist", configFile))
	}
	err = globalConf.Load(configFile)
	if err != nil {
		panic(fmt.Sprintf("load config file(%v) failed, err(%v)", configFile, err))
	}
	fmt.Println("load config" + configFile)
	fmt.Println("debug is ", globalConf.Debug)

	filename := fmt.Sprintf("%s/%s", globalConf.Log.Dir, globalConf.AppName)
	if FlagLogPidName != "" {
		filename += "." + FlagLogPidName
	}
	fmt.Println("xzap log file: ", filename)
	if globalConf.Log.Debug {
		xzap.InitLog(filename, xzap.WithLevel(xzap.Debug))
	} else {
		xzap.InitLog(filename)
	}
}

func GetGlobalConfig() *Config {
	return &globalConf
}

type Port struct {
	Prometheus int `toml:"prometheus" json:"prometheus"`
}
type Jaeger struct {
	Disable     bool   `toml:"disable" json:"disable"`
	AgentPort   int    `toml:"agent_port" json:"agent_port"`
	Payload     bool   `toml:"payload" json:"payload"`
	ServiceName string `toml:"service_name" json:"service_name"`
}

type Config struct {
	Debug         bool        `toml:"debug" json:"debug"`
	AppName       string      `toml:"app_name" json:"app_name"`
	Port          Port        `toml:"port" json:"port"`
	Jaeger        Jaeger      `toml:"jaeger" json:"jaeger"`
	EnvoyAddrTest string      `toml:"envoy_addr_test" json:"envoy_addr_test"`
	Log           Log         `toml:"log" json:"log"`
	ProtoFile     ProtoFile   `toml:"protofile" json:"protofile"`
	Mongo         Mongo       `toml:"mongo" json:"mongo"`
	MongoOps      Mongo       `toml:"mongo_ops" json:"mongo_ops"`
	Auth          Auth        `toml:"auth" json:"auth"`
	CiBot         CiBot       `toml:"cibot" json:"cibot"`
	Wechat        Wechat      `toml:"WECHAT" json:"wechat"`
	WechatGroup   WechatGroup `toml:"wechatgroup" json:"wechatgroup"`
	Flag          Flag        `toml:"flag" json:"flag"`
	Tapd          []TAPD      `toml:"tapd" json:"tapd"`
}

func (c *Config) Load(filePath string) error {
	_, err := toml.DecodeFile(filePath, &globalConf)
	return err
}

type Flag struct {
	PrivateDeploy bool   `toml:"private_deploy" json:"private_deploy"`
	Version       string `toml:"version" json:"version"`
}
type Log struct {
	Dir     string `toml:"dir" json:"dir"`
	Runtime bool   `toml:"runtime" json:"runtime"`
	Debug   bool   `toml:"debug" json:"debug"`
}

type ProtoFile struct {
	RootPath string `toml:"root_path" json:"root_path"`
}

type Mongo struct {
	Uri         string `toml:"uri" json:"uri"`
	Name        string `toml:"name" json:"name"`
	DBName      string `toml:"dbname" json:"dbname"`
	Password    string `toml:"password" json:"password"`
	Host        string `toml:"host" json:"host"`
	Port        int    `toml:"port" json:"port"`
	MaxPoolSize int    `toml:"max_pool_size" json:"max_pool_size"`
	MinPoolSize int    `toml:"min_pool_size" json:"min_pool_size"`
	MaxIdleTime int    `toml:"max_idle_time" json:"max_idle_time"`
}

type Auth struct {
	TestSecret string `toml:"test_secret" json:"test_secret"`
	TestPubKey string `toml:"test_pub_key" json:"test_pub_key"`
	ProdSecret string `toml:"prod_secret" json:"prod_secret"`
	ProdPubKey string `toml:"prod_pub_key" json:"prod_pub_key"`
}

type CiBot struct {
	Host string `toml:"host" json:"host"`
}

type Wechat struct {
	WechatUrl     string `toml:"wechat_url" json:"wechat_url"`
	WechatSecret  string `toml:"wechat_secret" json:"wechat_secret"`
	WechatPubkey  string `toml:"wechat_pubkey" json:"wechat_pubkey"`
	WechatGroupId string `toml:"wechat_group_id" json:"wechat_group_id"`
	GroupSecret   string `toml:"group_secret" json:"group_secret"`
	GroupKey      string `toml:"group_key" json:"group_key"`
}

type WechatGroup struct {
	Host string `toml:"host" json:"host"`
}

type TAPD struct {
	ApiUser     string `toml:"api_user" json:"api_user"`
	ApiPassword string `toml:"api_password" json:"api_password"`
	Name        string `toml:"name" json:"name"`
	WorkId      string `toml:"work_id" json:"work_id"`
}

//type TAPDProject struct {
//
//}
