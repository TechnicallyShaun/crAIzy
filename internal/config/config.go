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
	AIsFile    = "ais.yaml" // Deprecated: Use AgentsFile instead
	AgentsFile = "agents.yaml"
)

// Config represents the crAIzy configuration
type Config struct {
	ProjectName string   `yaml:"project_name"`
	AIs         []AISpec `yaml:"ais,omitempty"`
	Agents      []Agent  `yaml:"agents,omitempty"`
}

// AISpec defines an AI configuration (deprecated, use Agent)
type AISpec struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Options map[string]string `yaml:"options,omitempty"`
}

// Agent defines a CLI-based AI agent
type Agent struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
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

	// Create default agents config (CLI-based)
	defaultAgents := []Agent{
		{
			Name:    "Claude",
			Command: "claude --dangerously-skip-permissions",
		},
		{
			Name:    "Copilot",
			Command: "copilot --allow-all-tools",
		},
		{
			Name:    "Aider",
			Command: "aider",
		},
	}

	agentsPath := filepath.Join(craizyDir, AgentsFile)
	if err := saveAgents(agentsPath, defaultAgents); err != nil {
		return fmt.Errorf("failed to save agents config: %w", err)
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

	// Load AIs (for backward compatibility)
	aisPath := filepath.Join(ConfigDir, AIsFile)
	ais, err := loadAIs(aisPath)
	if err == nil {
		cfg.AIs = ais
	}

	// Load Agents
	agentsPath := filepath.Join(ConfigDir, AgentsFile)
	agents, err := loadAgents(agentsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load agents: %w", err)
	}
	cfg.Agents = agents

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

func saveAgents(path string, agents []Agent) error {
	data, err := yaml.Marshal(map[string][]Agent{"agents": agents})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadAgents(path string) ([]Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result struct {
		Agents []Agent `yaml:"agents"`
	}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Agents, nil
}

// AddAgent adds a new agent to the configuration
func AddAgent(name, command string) error {
	if !IsInitialized() {
		return fmt.Errorf("not in a crAIzy project")
	}

	agentsPath := filepath.Join(ConfigDir, AgentsFile)
	agents, err := loadAgents(agentsPath)
	if err != nil {
		// If file doesn't exist, start with empty list
		agents = []Agent{}
	}

	// Check if agent with this name already exists
	for _, agent := range agents {
		if agent.Name == name {
			return fmt.Errorf("agent with name '%s' already exists", name)
		}
	}

	// Add new agent
	agents = append(agents, Agent{
		Name:    name,
		Command: command,
	})

	return saveAgents(agentsPath, agents)
}

// RemoveAgent removes an agent from the configuration
func RemoveAgent(name string) error {
	if !IsInitialized() {
		return fmt.Errorf("not in a crAIzy project")
	}

	agentsPath := filepath.Join(ConfigDir, AgentsFile)
	agents, err := loadAgents(agentsPath)
	if err != nil {
		return err
	}

	// Find and remove agent
	found := false
	newAgents := make([]Agent, 0)
	for _, agent := range agents {
		if agent.Name != name {
			newAgents = append(newAgents, agent)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("agent with name '%s' not found", name)
	}

	return saveAgents(agentsPath, newAgents)
}

// ListAgents returns all configured agents
func ListAgents() ([]Agent, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("not in a crAIzy project")
	}

	agentsPath := filepath.Join(ConfigDir, AgentsFile)
	return loadAgents(agentsPath)
}
