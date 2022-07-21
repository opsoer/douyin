package main

import (
	"fmt"
	"net"

	"dy_uer_srv/global"
	"dy_uer_srv/handle"
	"dy_uer_srv/initialize"
	"dy_uer_srv/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.FeedSrv()
	zap.S().Info("user_srv 配置信息：%v", global.ServerConfig)
	zap.S().Info("项目初始化成功")
	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handle.UserServer{})
	addr := fmt.Sprintf("%v:%d", global.ServerConfig.Host, global.ServerConfig.Port)
	zap.S().Infof("user-srv启动地址%v", addr)
	lis, err := net.Listen("tcp", addr)
	//lis, err := net.Listen("tcp","127.0.0.1:8089")
	if err != nil {
		panic("failed to listen:" + err.Error())
	}
	server.Serve(lis)
}
