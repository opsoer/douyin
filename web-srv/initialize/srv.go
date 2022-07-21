package initialize

import (
	"dy_web_srv/global"
	"dy_web_srv/proto"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitSrv() {
	conn, err := grpc.Dial(fmt.Sprintf("%v:%d", global.SerConfig.UserSrvInfo.Host,
		global.SerConfig.UserSrvInfo.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Panic(err)
		panic(fmt.Sprintln("conn user srv err: ", err))
	}
	global.UserSrvCli = proto.NewUserClient(conn)

	conn2, err := grpc.Dial(fmt.Sprintf("%v:%d", global.SerConfig.FeedSrvInfo.Host,
		global.SerConfig.FeedSrvInfo.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Panic(err)
		panic(fmt.Sprintln("conn feed srv err: ", err))
	}
	global.FeedSrvCli = proto.NewFeedClient(conn2)
}
