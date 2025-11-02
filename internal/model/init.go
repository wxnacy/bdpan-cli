// Package model is the initial database driver and define the data structures corresponding to the tables.
package model

import (
	"os"
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
	db       *gorm.DB
	once1    sync.Once
	dbConfig = Sqlite{
		Connmaxlifetime: 60,
		Enablelog:       false,
		Maxidleconns:    10,
		Maxopenconns:    100,
	}
)

type Sqlite struct {
	Connmaxlifetime int  `yaml:"connmaxlifetime" json:"connmaxlifetime"`
	Enablelog       bool `yaml:"enablelog" json:"enablelog"`
	Maxidleconns    int  `yaml:"maxidleconns" json:"maxidleconns"`
	Maxopenconns    int  `yaml:"maxopenconns" json:"maxopenconns"`
}

// InitDB connect database
func InitDB() {
	InitSqlite()
	// switch strings.ToLower(config.Get().Database.Driver) {
	// case ggorm.DBDriverSqlite:
	// InitSqlite()
	// default:
	// panic("InitDB error, unsupported database driver: " + config.Get().Database.Driver)
	// }
}

// InitSqlite connect sqlite
func InitSqlite() {
	opts := []ggorm.Option{
		ggorm.WithMaxIdleConns(dbConfig.Maxidleconns),
		ggorm.WithMaxOpenConns(dbConfig.Maxopenconns),
		ggorm.WithConnMaxLifetime(time.Duration(dbConfig.Connmaxlifetime) * time.Minute),
	}
	if dbConfig.Enablelog {
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
		logFile, err := os.OpenFile(config.GetLogFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
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
	dbFile := utils.AdaptiveSqlite(config.GetDBFile())
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
