package initial

import (
	"time"

	"github.com/wxnacy/bdpan-cli/configs"
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
	if tools.FileExists(configPath) {
		configFile := configs.Path(configPath)
		errConfig = config.InitConfig(configFile)
	} else {
		errConfig = config.InitConfigByCode()
	}
	if errConfig != nil {
		panic("init config error: " + errConfig.Error())
	}
}
