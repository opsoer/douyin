package global

import (
	"dy_feed_srv/config"
	"dy_feed_srv/proto"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
	ServerConfig config.ServerConfig
	UserSrvCli proto.UserClient
)
