package config

import (
	"bytes"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	configYml = []byte(`
app:
  name: "bdpan"
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
	return nil
}

func InitConfigByCode() error {
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(configYml))
	return viper.Unmarshal(&config)
}

func GetConfigPath() string {
	return "/Users/wxnacy/Documents/Configs/bdpan-cli/config/bdpan-cli.yml"
}
