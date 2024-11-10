package initial

import (
	"github.com/wxnacy/bdpan-cli/configs"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/go-tools"
)

var (
	version            string
	configFile         string
	enableConfigCenter bool
)

// InitApp initial app configuration
func InitApp() {
	initConfig()

	// initial logger
	logger.Init(config.Get().Logger.LogFileConfig.Filename)
	logger.Infof("Init Config %#v", config.Get())
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
