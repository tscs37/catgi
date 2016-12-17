package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Backend  DriverConfig `json:"backend"`
	Index    DriverConfig `json:"index"`
	HMACKey  string       `json:"jwtkey"`
	Users    []UserConfig `json:"users"`
	HTTPConf HTTPConfig   `json:"http"`
	LogLevel string       `json:"loglevel"`
}

type HTTPConfig struct {
	Port     uint16 `json:"port"`
	ListenOn string `json:"listen"`
}

type DriverConfig struct {
	Name   string                 `json:"driver"`
	Params map[string]interface{} `json:"params"`
}

type UserConfig struct {
	Username string `json:"username"`
	PassHash string `json:"password"`
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
	return c, nil
}
