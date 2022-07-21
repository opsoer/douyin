package global

import (
	"dy_uer_srv/config"
	"dy_uer_srv/proto"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
	ServerConfig config.ServerConfig
	FeedSrvCli proto.FeedClient
)
