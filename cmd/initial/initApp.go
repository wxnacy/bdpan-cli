package initial

import (
	"time"

	"github.com/wxnacy/bdpan-cli/configs"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/go-tools"
)

// InitApp initial app configuration
func InitApp(req *dto.GlobalReq) {
	begin := time.Now()
	initConfig(req)

	// initial logger
	logger.Init()
	logger.Debugf("Init Config %#v time used %v", config.Get(), time.Now().Sub(begin))
}

func initConfig(req *dto.GlobalReq) {
	var configPath string
	var err error
	if req.Config != "" {
		configPath = req.Config
	} else {
		configPath, err = config.GetDefaultConfigPath()
		if err != nil {
			panic("get config path error: " + err.Error())
		}
	}
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
