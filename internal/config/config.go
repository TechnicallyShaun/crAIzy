package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigDir  = ".craizy"
	ConfigFile = "config.yaml"
	AIsFile    = "ais.yaml"
)

// Config represents the crAIzy configuration
type Config struct {
	ProjectName string   `yaml:"project_name"`
	AIs         []AISpec `yaml:"ais,omitempty"`
}

// AISpec defines an AI configuration
type AISpec struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Options map[string]string `yaml:"options,omitempty"`
}

// InitProject creates a new crAIzy project
func InitProject(name string) error {
	// Create project directory
	if err := os.MkdirAll(name, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create .craizy directory
	craizyDir := filepath.Join(name, ConfigDir)
	if err := os.MkdirAll(craizyDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .craizy directory: %w", err)
	}

	// Create default config
	cfg := Config{
		ProjectName: name,
	}

	configPath := filepath.Join(craizyDir, ConfigFile)
	if err := saveConfig(configPath, &cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Create default AIs config
	defaultAIs := []AISpec{
		{
			Name:    "GPT-4",
			Command: "openai-cli chat --model gpt-4",
			Options: map[string]string{
				"api_key": "$OPENAI_API_KEY",
			},
		},
		{
			Name:    "Claude",
			Command: "anthropic-cli chat --model claude-3-opus",
			Options: map[string]string{
				"api_key": "$ANTHROPIC_API_KEY",
			},
		},
		{
			Name:    "Local LLaMA",
			Command: "ollama run llama2",
			Options: map[string]string{},
		},
	}

	aisPath := filepath.Join(craizyDir, AIsFile)
	if err := saveAIs(aisPath, defaultAIs); err != nil {
		return fmt.Errorf("failed to save AIs config: %w", err)
	}

	return nil
}

// IsInitialized checks if current directory is a crAIzy project
func IsInitialized() bool {
	_, err := os.Stat(ConfigDir)
	return err == nil
}

// Load loads the configuration from the current directory
func Load() (*Config, error) {
	configPath := filepath.Join(ConfigDir, ConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Load AIs
	aisPath := filepath.Join(ConfigDir, AIsFile)
	ais, err := loadAIs(aisPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load AIs: %w", err)
	}
	cfg.AIs = ais

	return &cfg, nil
}

func saveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func saveAIs(path string, ais []AISpec) error {
	data, err := yaml.Marshal(map[string][]AISpec{"ais": ais})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadAIs(path string) ([]AISpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result struct {
		AIs []AISpec `yaml:"ais"`
	}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.AIs, nil
}
