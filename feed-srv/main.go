package main

import (
	"dy_feed_srv/global"
	"dy_feed_srv/handle"
	"dy_feed_srv/initialize"
	"dy_feed_srv/proto"

	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.UserSrv()
	zap.S().Info("user_srv 配置信息：%v", global.ServerConfig)
	zap.S().Info("项目初始化成功")
	server := grpc.NewServer()
	proto.RegisterFeedServer(server, &handle.FeedSrv{})
	addr := fmt.Sprintf("%v:%d", global.ServerConfig.Host, global.ServerConfig.Port)
	zap.S().Infof("feed-srv启动地址%v", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen:" + err.Error())
	}
	server.Serve(lis)
}
