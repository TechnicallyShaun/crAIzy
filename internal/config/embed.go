package config

import (
	_ "embed"
	"path/filepath"
)

//go:embed default_agents.yml
var DefaultAgentsYML []byte

const (
	// CraizyDir is the directory name for crAIzy configuration and data.
	CraizyDir = ".craizy"

	// AgentsFileName is the name of the agents configuration file.
	AgentsFileName = "AGENTS.yml"
)

// AgentsPath returns the path to the agents configuration file for a given work directory.
func AgentsPath(workDir string) string {
	return filepath.Join(workDir, CraizyDir, AgentsFileName)
}

// CraizyDirPath returns the path to the .craizy directory for a given work directory.
func CraizyDirPath(workDir string) string {
	return filepath.Join(workDir, CraizyDir)
}
