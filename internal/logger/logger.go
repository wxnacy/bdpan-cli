package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/go-bdpan"
)

func GetLogger() *logrus.Logger {
	return bdpan.GetLogger()
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Infoln(args ...interface{}) {
	GetLogger().Infoln(args...)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}
