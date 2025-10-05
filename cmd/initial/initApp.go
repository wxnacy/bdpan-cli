package initial

import (
	"time"

	"github.com/wxnacy/bdpan-cli/configs"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/go-tools"
)

// InitApp initial app configuration
func InitApp() {
	begin := time.Now()
	initConfig()

	// initial logger
	logger.Init()
	logger.Debugf("Init Config %#v time used %v", config.Get(), time.Now().Sub(begin))
}

func initConfig() {
	configPath := config.GetConfigPath()
	var err error
	if tools.FileExists(configPath) {
		configFile := configs.Path(configPath)
		err = config.InitConfig(configFile)
	} else {
		err = config.InitConfigByCode()
	}
	if err != nil {
		panic("init config error: " + err.Error())
	}
}
