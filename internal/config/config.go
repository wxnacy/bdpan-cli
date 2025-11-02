package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/wxnacy/bdpan-cli/internal/utils"
)

var (
	configPath string
	configYml  = []byte(`
app:
    name: bdpan
    scope: basic,netdisk
data_dir: "~/.local/share/bdpan"
`)
	initOnce sync.Once
)

// 初始化配置
// 功能需求:
// - 如果 config 变量有值，直接返回
// - 初始化的过程使用 sync.Once 保证只做一次
// - 先通过 configYml 加载配置 config
// - 如果 configFile 地址存在
//   - 如果内容不是 Yml 配置，则报错
//   - 是配置文件，则对 config 进行合并，字段值以文件中优先
//
// - *_dir/*_file 后缀的字段，使用 homedir.Expand 做 ~/ 地址开头解析
// - 如果 DataDir 为空，需要使用 utils.GetUserDataRoot 获取默认值
//
// 单元测试见本包 _test
func Init(configFile string) error {
	if config != nil {
		return nil
	}
	var initErr error
	initOnce.Do(func() {
		vip := viper.New()
		// 1) 载入内置默认配置
		vip.SetConfigType("yaml")
		if err := vip.ReadConfig(bytes.NewBuffer(configYml)); err != nil {
			initErr = err
			return
		}

		// 2) 合并外部配置（若存在）
		if configFile != "" {
			// 如果文件存在则尝试合并
			if info, err := os.Stat(configFile); err == nil && !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(configFile))
				if ext != ".yml" && ext != ".yaml" {
					initErr = fmt.Errorf("config file must be .yml or .yaml, got: %s", ext)
					return
				}
				vip.SetConfigFile(configFile)
				if err := vip.MergeInConfig(); err != nil {
					initErr = err
					return
				}
			} else if err != nil && !os.IsNotExist(err) {
				initErr = err
				return
			}
		}

		// 3) 反序列化到全局 config
		cfg := &Config{}
		if err := vip.Unmarshal(&cfg); err != nil {
			initErr = err
			return
		}

		// 4) DataDir 默认与路径展开（使用 utils.GetUserDataRoot）
		if strings.TrimSpace(cfg.DataDir) == "" {
			if root, err := utils.GetUserDataRoot(); err == nil {
				cfg.DataDir = root
			} else {
				initErr = err
				return
			}
		}
		if expanded, err := homedir.Expand(cfg.DataDir); err == nil {
			cfg.DataDir = expanded
		} else {
			initErr = err
			return
		}

		// 5) 设置全局
		config = cfg
	})
	return initErr
}

func InitConfig(configFile string) error {
	if config != nil {
		confFileAbs, err := filepath.Abs(configFile)
		if err != nil {
			return err
		}

		filePathStr, filename := filepath.Split(confFileAbs)
		ext := strings.TrimLeft(path.Ext(filename), ".")
		filename = strings.ReplaceAll(filename, "."+ext, "") // excluding suffix names

		viper.AddConfigPath(filePathStr) // path
		viper.SetConfigName(filename)    // file name
		viper.SetConfigType(ext)         // get the configuration type from the file name
		err = viper.ReadInConfig()
		if err != nil {
			return err
		}

		err = viper.Unmarshal(&config)
		if err != nil {
			return err
		}
	}
	err := FormatConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func InitConfigByCode() error {
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(configYml))
	err := viper.Unmarshal(&config)
	if err != nil {
		return err
	}
	err = FormatConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func FormatConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config not found")
	}
	var path string
	var err error
	path, err = homedir.Expand(config.DataDir)
	if err != nil {
		return err
	}
	config.DataDir = path
	return nil
}

func ReInitConfig() error {
	config = nil
	configPath := GetConfigPath()
	return InitConfig(configPath)
}

func SetConfigPath(path string) {
	// logger.Debugf("set config path: %s", path)
	configPath = path
}

func GetConfigPath() string {
	// logger.Debugf("config path: %s", configPath)
	return configPath
}

func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bdpan-cli", "config.yml"), nil
}

func GetLogFile() string {
	return filepath.Join(Get().DataDir, "log", "bdpan.log")
}

func GetDBFile() string {
	return filepath.Join(Get().DataDir, "bdpan.db?_busy_timeout=2&check_same_thread=false&cache=shared&mode=rwc")
}

// 获取缓存目录
func GetCacheDir() string {
	return filepath.Join(Get().DataDir, "cache")
}
