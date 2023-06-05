package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户对象
type User struct {
	Uid int64 `json:"uid" gorm:"column:uid;primary_key"`

	PwdHash string `gorm:"column:pwd_hash"`
	Salt    string `gorm:"column:salt"`

	Name   string `gorm:"column:name"`
	Role   uint32 `gorm:"column:role"`
	Perm   uint64 `gorm:"column:perm"`
	Status uint8  `gorm:"column:status"`

	CT time.Time `gorm:"column:ct"`
	UT time.Time `gorm:"column:ut"`
}

func (User) TableName() string {
	return "users"
}

func GetUser(db *gorm.DB, uid int64) (*User, error) {
	var user User
	if err := db.Where("uid=?", uid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByName(db *gorm.DB, name string) (*User, error) {
	var user User
	if err := db.Where("name=?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, uid int64, change map[string]interface{}) error {
	return db.Model(&User{}).Where("uid=?", uid).Updates(change).Error
}
