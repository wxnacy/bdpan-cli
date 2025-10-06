package config

import (
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/wxnacy/go-tools"
)

func (c *Credential) IsNil() bool {
	if c.AppID == 0 ||
		c.AppKey == "" ||
		c.SecretKey == "" {
		return true
	}
	return false
}

func SaveCredential(c Credential) error {
	viper.Set("credential", c)
	path := GetConfigPath()
	configDir := filepath.Dir(path)
	err := tools.DirExistsOrCreate(configDir)
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(path)
}
