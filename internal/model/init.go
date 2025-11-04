// Package model is the initial database driver and define the data structures corresponding to the tables.
package model

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/wxnacy/bdpan-cli/internal/config"
	log "github.com/wxnacy/bdpan-cli/internal/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	InitSqlite()
}

// InitSqlite 初始化 SQLite 数据库连接
//
// 实现逻辑（与代码顺序一致）:
// 1. 从 config.GetDBFile() 获取 DSN，并展开用户目录（~/）
// 2. 确保数据库文件所在目录存在（权限 0755）
// 3. 使用 gorm.io/driver/sqlite 打开数据库（禁用 SQL 日志，启用 PrepareStmt）
// 4. 配置连接池（SetMaxOpenConns=2, SetMaxIdleConns=2, SetConnMaxLifetime=0）
// 5. 执行 PRAGMA 兜底（journal_mode=WAL, synchronous=NORMAL, busy_timeout=5000, foreign_keys=ON, temp_store=MEMORY）
// 6. 自动迁移（AutoMigrate）模型表结构
//
// 参数: 无
// 返回: 无（失败时 panic）
func InitSqlite() {
	// 1. 获取 DSN（已包含 WAL 等参数）
	dbFile := config.GetDBFile()
	dbFile, _ = homedir.Expand(dbFile)

	// 2. 确保数据库目录存在
	dbDir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		panic("InitSqlite: failed to create db directory: " + err.Error())
	}

	// 3. 打开数据库
	var err error
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent), // 禁用 SQL 日志
		PrepareStmt: true,                                  // 启用预编译语句
	})
	if err != nil {
		panic("InitSqlite: failed to open database: " + err.Error())
	}

	// 4. 配置连接池（适配 SQLite）
	sqlDB, err := db.DB()
	if err != nil {
		panic("InitSqlite: failed to get sqlDB: " + err.Error())
	}
	sqlDB.SetMaxOpenConns(2)    // WAL 模式下 2 个连接足够
	sqlDB.SetMaxIdleConns(2)    // 保持 2 个空闲连接
	sqlDB.SetConnMaxLifetime(0) // 不主动回收

	// 5. 执行 PRAGMA 兜底（确保 DSN 参数生效）
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA temp_store=MEMORY;",
	}
	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			log.Infof("InitSqlite: failed to execute %s: %v", pragma, err)
		}
	}

	// 6. 自动迁移表结构
	begin := time.Now()
	if err := db.AutoMigrate(&UploadHistory{}, &File{}, &Quick{}); err != nil {
		panic("InitSqlite: AutoMigrate failed: " + err.Error())
	}
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

// CloseDB 关闭数据库连接
//
// 实现逻辑:
// 1. 获取底层 *sql.DB
// 2. 调用 Close() 关闭连接池
// 3. WAL 模式下会自动执行 checkpoint
//
// 返回: error
func CloseDB() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
