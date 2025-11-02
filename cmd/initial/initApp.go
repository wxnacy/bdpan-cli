package initial

import (
	"path/filepath"
	"time"

	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/go-tools"
)

// InitApp initial app configuration
func InitApp() {
	begin := time.Now()
	initConfig()

	// initial logger
	logger.Init()
	logger.Debugf("Init Config %#v time used %v", config.Get(), time.Since(begin))
}

func initConfig() {
	configPath := handler.GetRequest().GetConfigPath()
	config.SetConfigPath(configPath)

	var errConfig error
	errConfig = config.Init(configPath)
	if errConfig != nil {
		panic("init config error: " + errConfig.Error())
	}
	tools.DirExistsOrCreate(config.GetCacheDir())
	tools.DirExistsOrCreate(filepath.Dir(config.GetLogFile()))
}
