package initialize

import (
	"dy_uer_srv/global"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitConfig() {
	v := viper.New()
	//文件的路径如何设置
	v.SetConfigFile("user_srv.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&global.ServerConfig); err != nil {
		zap.S().Infof("读取配置错误%v", err)
	}

}
