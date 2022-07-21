package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
)

type ComIdList []int32

func (c ComIdList) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ComIdList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &c)
}

type Video struct {
	Id             int32
	CreateTime     int64
	UpdateTime     int64
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	UserID         int32
	PlayUrl        string
	CoverUrl       string
	Title          string
	Favorite_count int32
	Comment_count  int32
	CommentList    ComIdList
}

type Comment struct {
	gorm.Model
	UserId     int32
	Content    string
}
