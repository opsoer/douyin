package initialize

import (
	"dy_uer_srv/global"
	"dy_uer_srv/proto"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func FeedSrv (){
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.FeedSrvInfo.Host,
		global.ServerConfig.FeedSrvInfo.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Error(err)
		return
	}
	global.FeedSrvCli = proto.NewFeedClient(conn)
}

