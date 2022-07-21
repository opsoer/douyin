package main

import (
	"dy_feed_srv/global"
	"dy_feed_srv/initialize"
	model "dy_feed_srv/modle"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	serConfig := global.ServerConfig.MysqlInfo
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		serConfig.User, serConfig.Password, serConfig.Host, serConfig.Port, serConfig.Name)
	zap.S().Infof(dsn)
	//	dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	//	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	global.DB.AutoMigrate(&model.Video{})
	global.DB.AutoMigrate(&model.Comment{})
	//global.DB.Create(&model.Video{
	//	Id:             1100,
	//	CreateTime:     time.Now().Unix(),
	//	UpdateTime:     time.Now().Unix(),
	//	UserID:         1,
	//	PlayUrl:        "www.playUrl1001.com",
	//	CoverUrl:       "www.coverUrl1001.com",
	//	Title:          "第5个视频,并且有评论",
	//	CommentList: []int32{1,2,3,4,5,6,7,8,9},
	//})
	//videoList := make([]model.Video, 0)
	//global.DB.Where("update_time >= ?", 1653379096).Find(&videoList)
	//for _, video := range videoList {
	//	fmt.Println(video)
	//}
	//videoList := make([]model.Video, 0)
	//if result := global.DB.Where(&model.Video{Id: 1001}).Find(&videoList); result.Error != nil {
	//	fmt.Println("获取评论列表失败：", result.Error)
	//	return
	//}
	//for _, video := range videoList {
	//	for _, comId := range video.CommentList {
	//		fmt.Println(comId)
	//	}
	//}



}
