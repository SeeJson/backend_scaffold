package app

import (
	"context"
	"gorm.io/gorm"
	"time"

	"github.com/SeeJson/backend_scaffold/model"
)

type OpWriter struct {
	db *gorm.DB
}

func NewOpWriter(db *gorm.DB) *OpWriter {
	return &OpWriter{db: db}
}

func (ow *OpWriter) WriteOpLog(ctx context.Context, op model.OperationLog) {
	now := time.Now()
	op.CT = now
	op.UT = now
	if op.Value == "" {
		op.Value = "{}"
	}

	if err := ow.db.Create(&op).Error; err != nil {
		GetLogger(ctx).WithField("err", err).WithField("op", op).Error("write operation_log error")
	}
}
