package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Backend DriverConfig `json:"backend"`
	Index   DriverConfig `json:"index"`
}

type DriverConfig struct {
	Name   string                 `json:"driver"`
	Params map[string]interface{} `json:"params"`
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
