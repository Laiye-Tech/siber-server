module api-test

replace github.com/robfig/cron/v3 v3.0.1 => github.com/pradykaushik/cron/v3 v3.0.2-0.20200114011641-62e4aea507a6 // indirect

go 1.13

require (
	git.laiye.com/laiye-backend-repos/go-utils/xzap v0.1.6
	git.laiye.com/laiye-backend-repos/im-saas-protos-golang v5.21.30+incompatible
	github.com/BurntSushi/toml v0.3.1
	github.com/astaxie/beego v1.10.1
	github.com/bxcodec/faker/v3 v3.5.0
	github.com/deckarep/golang-set v1.7.1
	github.com/fullstorydev/grpcurl v1.8.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.11.3
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/jhump/protoreflect v1.6.1
	github.com/jinzhu/gorm v1.9.11
	github.com/lib/pq v1.3.0 // indirect
	github.com/mwitkow/go-proto-validators v0.2.0 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/opentracing/opentracing-go v1.1.0
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.6.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/tidwall/gjson v1.6.0
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	go.mongodb.org/mongo-driver v1.1.2
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc v1.30.0
	gopkg.in/yaml.v2 v2.3.0 // indirect

)
