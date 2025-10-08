// Package model is the initial database driver and define the data structures corresponding to the tables.
package model

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-dev-frame/sponge/pkg/ggorm"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"github.com/mitchellh/go-homedir"
	"github.com/wxnacy/bdpan-cli/internal/config"
	log "github.com/wxnacy/bdpan-cli/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// ErrCacheNotFound No hit cache

// ErrRecordNotFound no records found
var ErrRecordNotFound = gorm.ErrRecordNotFound

var (
	db    *gorm.DB
	once1 sync.Once
)

// InitDB connect database
func InitDB() {
	switch strings.ToLower(config.Get().Database.Driver) {
	case ggorm.DBDriverSqlite:
		InitSqlite()
	default:
		panic("InitDB error, unsupported database driver: " + config.Get().Database.Driver)
	}
}

// InitSqlite connect sqlite
func InitSqlite() {
	opts := []ggorm.Option{
		ggorm.WithMaxIdleConns(config.Get().Database.Sqlite.MaxIdleConns),
		ggorm.WithMaxOpenConns(config.Get().Database.Sqlite.MaxOpenConns),
		ggorm.WithConnMaxLifetime(time.Duration(config.Get().Database.Sqlite.ConnMaxLifetime) * time.Minute),
	}
	if config.Get().Database.Sqlite.EnableLog {
		l := logger.Get()
		// 配置日志输出到文件
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		logFile, err := os.OpenFile(config.Get().Logger.LogFileConfig.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
		if err != nil {
			panic(err)
		}
		l = l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(logFile),
				zapcore.DebugLevel,
			)
		}))
		opts = append(opts,
			ggorm.WithLogging(l),
			ggorm.WithLogRequestIDKey("request_id"),
		)
	}
	opts = append(opts, ggorm.WithEnableTrace())

	// if config.Get().App.EnableTrace {
	// opts = append(opts, ggorm.WithEnableTrace())
	// }

	var err error
	dbFile := utils.AdaptiveSqlite(config.Get().Database.Sqlite.DBFile)
	dbFile, _ = homedir.Expand(dbFile)
	db, err = ggorm.InitSqlite(dbFile, opts...)
	if err != nil {
		panic("InitSqlite error: " + err.Error())
	}

	// 初始化表格
	begin := time.Now()
	db.AutoMigrate(&UploadHistory{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&Quick{})
	log.Debugf("DB AutoMigrate time used %v", time.Since(begin))
}

// GetDB get db
func GetDB() *gorm.DB {
	if db == nil {
		once1.Do(func() {
			InitDB()
		})
	}

	return db
}

// CloseDB close db
func CloseDB() error {
	return ggorm.CloseDB(db)
}
