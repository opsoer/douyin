package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
)

type Videos []int32

func (v Videos) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *Videos) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &v)
}

type Follows []int32

func (f Follows) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *Follows) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &f)
}

type Followers []int32

func (f Followers) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *Followers) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &f)
}

type Favorites []int32

func (f Favorites) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *Favorites) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &f)
}

type User struct {
	gorm.Model
	Password       string `gorm:"type:varchar(100);not null"`
	Name           string `gorm:"type:varchar(20)"`
	Follow_count   int32
	Follower_count int32
	FollowerList   Followers
	FollowList     Follows
	VideoList      Videos
	FavList        Favorites
}
