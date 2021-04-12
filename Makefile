WORKING_DIR := $(shell pwd)
GOPATH=/var/lib/jenkins/go
PROPATH=/api-test
cc=docker run --rm -t -v ${WORKING_DIR}:${PROPATH} -v ${GOPATH}:/go -e GOPROXY=http://goproxy-inner.laiye.com golang:1.12.1
test:
	 echo 'test complete'

siber:
	 $(cc) /bin/sh -c 'cd ${PROPATH}&&go build -o api-test-siber main/main.go'

build:
	 echo "make build"