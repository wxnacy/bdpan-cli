package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wxnacy/bdpan"
)

func Init(logPath string) error {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
	if err != nil {
		return err
	}
	bdpan.Log.SetOutput(logFile)
	return nil
}

func SetLogLevel(level logrus.Level) {
	bdpan.Log.SetLevel(level)
}
