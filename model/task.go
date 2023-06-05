package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID  int64 `json:"id" gorm:"column:id;primary_key"`
	Uid int64 `json:"uid" gorm:"column:uid"`

	Status uint8 `gorm:"column:status"`

	Priority int8   `gorm:"column:priority"`
	Type     int8   `gorm:"column:type"`
	Input    string `gorm:"column:input"`
	Result   string `gorm:"column:result"`

	Note string `gorm:"column:note"`

	CT time.Time `gorm:"column:ct"`
	UT time.Time `gorm:"column:ut"`
}

func (Task) TableName() string {
	return "tasks"
}

func GetTask(db *gorm.DB, id int64) (*Task, error) {
	var v Task
	if err := db.Where("id=?", id).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func UpdateTask(db *gorm.DB, id int64, change map[string]interface{}) error {
	return db.Model(&Task{}).Where("id=?", id).Updates(change).Error
}
