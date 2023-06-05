package app

import (
	"context"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const ctxLoggerKey = "/logger"

func SetLogger(c *gin.Context, entry *log.Entry) {
	c.Set(ctxLoggerKey, entry)
}

func GetLogger(ctx context.Context) *log.Entry {
	logger, ok := ctx.Value(ctxLoggerKey).(*log.Entry)
	if ok && logger != nil {
		return logger
	}

	return log.NewEntry(log.StandardLogger())
}
