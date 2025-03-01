package config

import (
	"github.com/spf13/viper"
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
	return viper.WriteConfigAs(GetConfigPath())
}
