package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

func getCredentialPath() (string, error) {
	dataDir := Get().DataDir
	if dataDir == "" {
		panic("config.dataDir is empty")
	}
	p := filepath.Join(dataDir, "credential.json")
	return homedir.Expand(p)
}

func GetCredential() (*Credential, error) {
	p, err := getCredentialPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var a Credential
	err = json.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

type Credential struct {
	AppID     int    `yaml:"appId" json:"appId"`
	AppKey    string `yaml:"appKey" json:"appKey"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	SignKey   string `yaml:"signKey" json:"signKey"`
}

func (c *Credential) IsNil() bool {
	if c.AppID == 0 ||
		c.AppKey == "" ||
		c.SecretKey == "" {
		return true
	}
	return false
}

func SaveCredential(c Credential) error {
	p, err := getCredentialPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
