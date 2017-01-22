package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Configuration struct {
	Backend    DriverConfig `json:"backend"`
	Index      DriverConfig `json:"index"`
	HMACKey    string       `json:"jwtkey"`
	Pepper     string       `json:"password_pepper"`
	Users      []UserConfig `json:"users"`
	IgnoreAuth bool         `json:"ignore_login"`
	HTTPConf   HTTPConfig   `json:"http"`
	LogLevel   string       `json:"loglevel"`
	Piwik      PiwikConfig  `json:"piwik"`
}

type PiwikConfig struct {
	Enable       bool   `json:"enable"`
	Base         string `json:"base"`
	ID           string `json:"site_id"`
	IgnoreErrors bool   `json:"ignore_err"`
	Token        string `json:"admin_token"`
}
type HTTPConfig struct {
	Port     uint16 `json:"port"`
	ListenOn string `json:"listen"`
}

type DriverConfig struct {
	Name   string                 `json:"driver"`
	Params map[string]interface{} `json:"params"`
}

type AuthenticationType string

const (
	ATPasslib AuthenticationType = ""
	ATDropbox AuthenticationType = "dropbox"
)

type UserConfig struct {
	Username string             `json:"username"`
	PassHash string             `json:"password"`
	AuthType AuthenticationType `json:"authtype"`
}

func LoadConfig(path string) (Configuration, error) {
	var c = Configuration{}
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(dat, &c); err != nil {
		return c, err
	}
	if len(c.Pepper) > 0 && len(c.Pepper) != 32 {
		return c, errors.New("Pepper must be 32 characters")
	}
	return c, nil
}
