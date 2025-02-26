// Package model is the initial database driver and define the data structures corresponding to the tables.
package model

import (
	"strings"
	"sync"
	"time"

	"github.com/go-dev-frame/sponge/pkg/ggorm"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"github.com/mitchellh/go-homedir"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"gorm.io/gorm"
)

var (
	// ErrCacheNotFound No hit cache

	// ErrRecordNotFound no records found
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

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
		opts = append(opts,
			ggorm.WithLogging(logger.Get()),
			ggorm.WithLogRequestIDKey("request_id"),
		)
	}
	opts = append(opts, ggorm.WithEnableTrace())

	// if config.Get().App.EnableTrace {
	// opts = append(opts, ggorm.WithEnableTrace())
	// }

	var err error
	var dbFile = utils.AdaptiveSqlite(config.Get().Database.Sqlite.DBFile)
	dbFile, _ = homedir.Expand(dbFile)
	db, err = ggorm.InitSqlite(dbFile, opts...)
	if err != nil {
		panic("InitSqlite error: " + err.Error())
	}
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
