package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Agent struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type AgentsConfig struct {
	Agents []Agent `yaml:"agents"`
}

func LoadAgents(path string) ([]Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config AgentsConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config.Agents, nil
}
