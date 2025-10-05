package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/go-bdpan"
)

func Init() error {
	// logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
	// if err != nil {
	// return err
	// }
	// bdpan.SetOutput(logFile)
	// bdpan.SetOutputFile(logPath)
	SetLogFile()
	// SetLogLevel(logrus.DebugLevel)
	return nil
}

func SetLogLevel(level logrus.Level) {
	bdpan.SetLogLevel(level)
}

func SetLogFile() {
	bdpan.SetOutputFile(config.Get().Logger.LogFileConfig.Filename)
}

func ClearLogFile() {
	GetLogger().SetOutput(os.Stdout)
}
