package global

import (
	"dy_web_srv/config"
	"dy_web_srv/proto"
	"github.com/gin-gonic/gin"
)

var (
	SerConfig  config.ServerConfig
	Router     *gin.Engine
	UserSrvCli proto.UserClient
	FeedSrvCli proto.FeedClient
)
