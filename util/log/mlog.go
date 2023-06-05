package mlog

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	FileRoot string `mapstructure:"file_root"` // log文件所在目录
}

var cfg Config

func SetConfig(c Config) {
	cfg = c

	// Set Log
	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = log.InfoLevel
	}
	jsonFormatter := &log.JSONFormatter{
		TimestampFormat: "2006-01-02.15:04:05.000000",
	}
	logFileRoot := cfg.FileRoot
	if logFileRoot != "" {
		writer, _ := rotatelogs.New(cfg.FileRoot + "log-%Y%m%d.log")
		lfHook := lfshook.NewHook(lfshook.WriterMap{
			log.InfoLevel:  writer,
			log.DebugLevel: writer,
			log.ErrorLevel: writer,
		}, jsonFormatter)
		log.AddHook(lfHook)
	}
	log.SetLevel(logLevel)
	log.SetFormatter(jsonFormatter)
	log.SetReportCaller(true)
}
