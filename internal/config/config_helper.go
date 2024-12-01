package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Config ConfigModel

func LoadConfig() error {
	data, err := os.ReadFile("config/config.yml")
	if err != nil {
		return fmt.Errorf("unable to read config/config.yml file")
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("unable to ynmarshal yaml to config")
	}
	return nil
}
