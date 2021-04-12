package payload

import (
	"context"
	"time"

	"api-test/configs"
	"api-test/libs"

	"git.laiye.com/laiye-backend-repos/go-utils/xzap"
	"git.laiye.com/laiye-backend-repos/im-saas-protos-golang/cibot"
	"go.uber.org/zap"
)

type CiGrpcService struct {
	Conn libs.Connection
}

func NewCiService() *CiGrpcService {
	address := configs.GetGlobalConfig().CiBot.Host
	conn := libs.NewEnvoyResolverConnection(address)
	conn.Connect(context.Background())
	return &CiGrpcService{conn}
}

func SendPlanRes(ctx context.Context, msg string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	var CiServiceInstance = NewCiService()
	client := cibot.NewWorkWechatGroupBotServiceClient(CiServiceInstance.Conn.GetGrpcConnection(ctx))
	_, err := client.SendWorkWechatGroupByBot(ctx, &cibot.SendWorkWechatGroupByBotRequest{
		Type: cibot.WorkWechatGroupBotMessageType_TEXT,
		Body: &cibot.SendWorkWechatGroupByBotBody{
			Body: &cibot.SendWorkWechatGroupByBotBody_Text{Text: &cibot.WorkWechatGroupBotTextMessage{
				Content:             msg,
				MentionedList:       nil,
				MentionedMobileList: nil,
			}},
		},
		Url: configs.GetGlobalConfig().WechatGroup.Host,
	})
	if err != nil {
		xzap.Logger(ctx).Error("call cibot send wechat group msg", zap.Any("err", err))
	}
}
