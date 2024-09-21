package main

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Points []TriggerPoint `yaml:"points"`
	Host   string         `yaml:"host"`
	Port   int            `yaml:"port"`
}

func LoadConfig() (*Config, error) {
	f, err := os.Open("osctrigger_config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			config := Config{}
			return &config, nil
		}
		return nil, err
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(config Config) error {
	f, err := os.OpenFile("osctrigger_config.yaml", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	yamlFile, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, string(yamlFile))

	return err
}
