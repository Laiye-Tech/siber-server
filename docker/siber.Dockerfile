FROM registry-vpc.cn-beijing.aliyuncs.com/laiye-devops/golang:1.13 as builder

ENV GOPROXY=https://goproxy.cn

COPY . .
RUN go build -o api-test-siber main/main.go

FROM registry-vpc.cn-beijing.aliyuncs.com/laiye-devops/debian-kube:latest

COPY --from=builder /api-test-siber /home/works/program/
ADD docker/siber.supervisord.conf /etc/supervisord.conf