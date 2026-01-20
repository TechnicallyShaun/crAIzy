package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitProject(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	projectName := "test-project"
	projectPath := filepath.Join(tmpDir, projectName)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Test project initialization
	err := InitProject(projectName)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Verify project directory was created
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Errorf("Project directory not created")
	}

	// Verify .craizy directory was created
	craizyDir := filepath.Join(projectPath, ConfigDir)
	if _, err := os.Stat(craizyDir); os.IsNotExist(err) {
		t.Errorf(".craizy directory not created")
	}

	// Verify config file was created
	configFile := filepath.Join(craizyDir, ConfigFile)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("config.yaml not created")
	}

	// Verify AIs file was created
	aisFile := filepath.Join(craizyDir, AIsFile)
	if _, err := os.Stat(aisFile); os.IsNotExist(err) {
		t.Errorf("ais.yaml not created")
	}
}

func TestInitProjectDuplicateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-project"

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create directory first
	os.MkdirAll(projectName, 0755)

	// Should succeed even if directory exists
	err := InitProject(projectName)
	if err != nil {
		t.Fatalf("InitProject should handle existing directory: %v", err)
	}
}

func TestIsInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-project"

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Not initialized
	os.Chdir(tmpDir)
	if IsInitialized() {
		t.Errorf("Should not be initialized in empty directory")
	}

	// Initialize project
	InitProject(projectName)
	os.Chdir(filepath.Join(tmpDir, projectName))

	// Should be initialized
	if !IsInitialized() {
		t.Errorf("Should be initialized after init")
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-project"

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Initialize project
	err := InitProject(projectName)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Change to project directory
	os.Chdir(filepath.Join(tmpDir, projectName))

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify config
	if cfg.ProjectName != projectName {
		t.Errorf("Expected project name %s, got %s", projectName, cfg.ProjectName)
	}

	// Verify AIs were loaded
	if len(cfg.AIs) == 0 {
		t.Errorf("Expected AIs to be loaded")
	}

	// Verify default AIs
	aiNames := []string{"GPT-4", "Claude", "Local LLaMA"}
	if len(cfg.AIs) != len(aiNames) {
		t.Errorf("Expected %d default AIs, got %d", len(aiNames), len(cfg.AIs))
	}

	for i, expectedName := range aiNames {
		if i >= len(cfg.AIs) {
			break
		}
		if cfg.AIs[i].Name != expectedName {
			t.Errorf("Expected AI name %s, got %s", expectedName, cfg.AIs[i].Name)
		}
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Try to load without initializing
	_, err := Load()
	if err == nil {
		t.Errorf("Load should fail in non-initialized directory")
	}
}

func TestAISpec(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-project"

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Initialize and load
	InitProject(projectName)
	os.Chdir(filepath.Join(tmpDir, projectName))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify AI spec structure
	for _, ai := range cfg.AIs {
		if ai.Name == "" {
			t.Errorf("AI name should not be empty")
		}
		if ai.Command == "" {
			t.Errorf("AI command should not be empty")
		}
		// Options can be nil or empty map, both are valid
	}
}

func TestSaveAndLoadAIs(t *testing.T) {
	tmpDir := t.TempDir()
	aisFile := filepath.Join(tmpDir, "test-ais.yaml")

	// Test data
	testAIs := []AISpec{
		{
			Name:    "TestAI",
			Command: "test-command",
			Options: map[string]string{
				"key": "value",
			},
		},
	}

	// Save
	err := saveAIs(aisFile, testAIs)
	if err != nil {
		t.Fatalf("saveAIs failed: %v", err)
	}

	// Load
	loadedAIs, err := loadAIs(aisFile)
	if err != nil {
		t.Fatalf("loadAIs failed: %v", err)
	}

	// Verify
	if len(loadedAIs) != len(testAIs) {
		t.Errorf("Expected %d AIs, got %d", len(testAIs), len(loadedAIs))
	}

	if loadedAIs[0].Name != testAIs[0].Name {
		t.Errorf("Expected name %s, got %s", testAIs[0].Name, loadedAIs[0].Name)
	}

	if loadedAIs[0].Command != testAIs[0].Command {
		t.Errorf("Expected command %s, got %s", testAIs[0].Command, loadedAIs[0].Command)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	// Test data
	testCfg := &Config{
		ProjectName: "test",
	}

	// Save
	err := saveConfig(configFile, testCfg)
	if err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file not created")
	}
}
