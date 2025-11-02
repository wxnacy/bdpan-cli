package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	configPath string
	configYml  = []byte(`
app:
    name: bdpan
    scope: basic,netdisk
data_dir: "~/.local/share/bdpan"
`)
)

type Credentials []Credential

func InitConfig(configFile string) error {
	if config == nil {
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
