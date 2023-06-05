package utils

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Filename   string `mapstructure:"filename"`
	Level      string `mapstructure:"level"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	LocalTime  bool   `mapstructure:"local_time"`
	Compress   bool   `mapstructure:"compress"`
	Std        bool   `mapstructure:"std"`
}

func (c LogConfig) InitLog() {
	log.SetReportCaller(true)
	logLevel, err := log.ParseLevel(c.Level)
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
	log.SetFormatter(&log.JSONFormatter{})
	if log.GetLevel() == log.DebugLevel {
		// 设置日志格式为text格式
		// log.SetFormatter(&log.JSONFormatter{}) 设置为json格式
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	}

	if c.Filename == "" {
		return
	}

	ll := &lumberjack.Logger{
		Filename:   c.Filename,
		MaxSize:    c.MaxSize,
		MaxBackups: c.MaxBackups,
		MaxAge:     c.MaxAge,
		Compress:   c.Compress,
	}

	var mWriter io.Writer
	if c.Std {
		mWriter = io.MultiWriter(os.Stderr, ll)
	} else {
		mWriter = ll
	}
	log.SetOutput(mWriter)
}
