package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
)

type Access struct {
	AccessToken      string `yaml:"accessToken" json:"accessToken"`
	ExpiresIn        int    `yaml:"expiresIn" json:"expiresIn"`
	RefreshTimestamp int    `yaml:"refreshTimestamp" json:"refreshTimestamp"`
	RefreshToken     string `yaml:"refreshToken" json:"refreshToken"`
}

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

func getAccessPath() (string, error) {
	dataDir := Get().DataDir
	if dataDir == "" {
		panic("config.dataDir is empty")
	}
	p := filepath.Join(dataDir, "access.json")
	return homedir.Expand(p)
}

func SaveAccess(a Access) error {
	path, err := getAccessPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func GetAccess() (*Access, error) {
	path, err := getAccessPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a Access
	err = json.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func GetAccessToken() string {
	a, err := GetAccess()
	if err != nil {
		return ""
	}
	return a.AccessToken
}
