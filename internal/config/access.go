package config

import (
	"time"

	"github.com/spf13/viper"
)

func (a *Access) IsNil() bool {
	if a.AccessToken == "" ||
		a.ExpiresIn == 0 ||
		a.RefreshToken == "" {
		return true
	}
	return false
}

func (a *Access) IsExpired() bool {
	if a.IsNil() || int(time.Now().Unix()) > a.RefreshTimestamp {
		return true
	}
	return false
}

func SaveAccess(a Access) error {
	viper.Set("access", a)
	return viper.WriteConfigAs(GetConfigPath())
}

func GetAccessToken() string {
	return Get().Access.AccessToken
}
