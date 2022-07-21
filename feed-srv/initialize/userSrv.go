package initialize

import (
	"dy_feed_srv/proto"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"dy_feed_srv/global"
)

func UserSrv () {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrv.Host, global.ServerConfig.UserSrv.Port),
		grpc.WithInsecure())
	if err != nil {
		zap.S().Error(err)
		return
	}
	global.UserSrvCli = proto.NewUserClient(conn)

}
