package logger

import (
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

func Init() error {
	SetLogFile()
	return nil
}

func SetLogLevel(level logrus.Level) {
	bdpan.SetLogLevel(level)
}

func SetLogFile() {
	logPath := config.Get().Logger.LogFileConfig.Filename
	tools.DirExistsOrCreate(filepath.Dir(logPath))
	// 设置按日期分割日志，最多十个文件
	logf, err := rotatelogs.New(
		logPath+".%Y%m%d",
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithRotationCount(10),
	)
	if err != nil {
		GetLogger().Errorf("failed to create rotatelogs: %s", err)
		return
	}
	GetLogger().SetOutput(logf)
}

func ClearLogFile() {
	GetLogger().SetOutput(os.Stdout)
}
