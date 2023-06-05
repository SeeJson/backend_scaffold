package model

import "time"

// 审计日志
type OperationLog struct {
	API       string `gorm:"column:api"`
	Operation string `gorm:"column:operation"`

	Uid int64 `gorm:"column:uid"`

	Value string `gorm:"column:value"`

	CT time.Time `gorm:"column:ct"`
	UT time.Time `gorm:"column:ut"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
