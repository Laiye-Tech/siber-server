/**
* @Author: TongTongLiu
* @Date: 2020/5/26 7:51 下午
**/

package siberconst

const (
	GRPCProtocol string = "grpc"
	HTTPProtocol string = "http"
)

const (
	GraphQLMethod = "graphQL"
	GRPCMethod    = GRPCProtocol
	HTTPMethod    = HTTPProtocol
)

// 自定义的请求头
const (
	Siber       = "Siber"
	SiberAuth   = "SiberAuth"
	SiberPubkey = "pubkey"
	SiberSecret = "secret"
)

// 环境
const (
	EnvironmentTest  string = "test"
	EnvironmentDev   string = "dev"
	EnvironmentStage string = "stage"
	EnvironmentProd  string = "prod"
)
