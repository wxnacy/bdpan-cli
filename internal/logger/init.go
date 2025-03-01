package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/go-bdpan"
)

func Init(logPath string) error {
	// logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
	// if err != nil {
	// return err
	// }
	// bdpan.SetOutput(logFile)
	bdpan.SetOutputFile(logPath)
	return nil
}

func SetLogLevel(level logrus.Level) {
	bdpan.SetLogLevel(level)
}
