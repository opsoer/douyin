package main

import (
	"dy_web_srv/global"
	"dy_web_srv/initialize"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitSrv()
	global.Router = gin.Default()
	initialize.InitUserRouter()
	zap.S().Info("web服务配置信息", global.SerConfig)
	zap.S().Debugf("启动服务器启动地址%v:%d", global.SerConfig.Host, global.SerConfig.Port)
	if err := global.Router.Run(fmt.Sprintf("%v:%d", global.SerConfig.Host, global.SerConfig.Port)); err != nil {
		zap.S().Panic("启动失败:", err.Error())
	}
}