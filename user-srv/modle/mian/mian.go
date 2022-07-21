package main

import (
	model "dy_uer_srv/modle"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/dy_test?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	// 全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&model.User{})
	if err != nil {
		fmt.Println(err)
	}
	//videos := []model.Video{{1}, {2}, {3}}

	//user := model.User{
	//	Password:       "123456",
	//	Name:           "zr04",
	//}
	//_ = db.Create(&user)
//	fmt.Println(result.Error)
//	fmt.Println(result.RowsAffected)
//	us1 := model.User{}
//	db.Model(&model.User{}).First(&us1,2)
//	fmt.Println(us1)
}
