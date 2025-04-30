package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := []byte(`lunchMoneyApiKey: test-api-key`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config was loaded correctly
	if config.LunchMoneyAPIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", config.LunchMoneyAPIKey)
	}
}

func TestLoadConfigError(t *testing.T) {
	// Test loading a non-existent config file
	_, err := LoadConfig("non-existent-file.yaml")
	if err == nil {
		t.Errorf("Expected error when loading non-existent file, got nil")
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an invalid YAML file
	invalidPath := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := []byte(`invalid: yaml: content`)
	if err := os.WriteFile(invalidPath, invalidContent, 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	// Test loading an invalid config file
	_, err = LoadConfig(invalidPath)
	if err == nil {
		t.Errorf("Expected error when loading invalid YAML, got nil")
	}
}

func TestInitGlobalConfig(t *testing.T) {
	// Reset global config for testing
	configMutex.Lock()
	globalConfig = nil
	configLoaded = false
	configMutex.Unlock()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := []byte(`lunchMoneyApiKey: test-api-key`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test initializing the global config
	err = InitGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize global config: %v", err)
	}

	// Verify the global config was initialized correctly
	configMutex.RLock()
	defer configMutex.RUnlock()
	if !configLoaded {
		t.Errorf("Expected configLoaded to be true, got false")
	}
	if globalConfig == nil {
		t.Fatalf("Expected globalConfig to be non-nil, got nil")
	}
	if globalConfig.LunchMoneyAPIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", globalConfig.LunchMoneyAPIKey)
	}
}

func TestGetConfig(t *testing.T) {
	// Reset global config for testing
	configMutex.Lock()
	globalConfig = nil
	configLoaded = false
	configMutex.Unlock()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := []byte(`lunchMoneyApiKey: test-api-key`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Initialize the global config
	err = InitGlobalConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize global config: %v", err)
	}

	// Test getting the config
	config, err := GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	// Verify the config was retrieved correctly
	if config.LunchMoneyAPIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", config.LunchMoneyAPIKey)
	}
}

func TestGetLunchMoneyAPIKey(t *testing.T) {
	// Reset global config for testing
	configMutex.Lock()
	globalConfig = &Config{LunchMoneyAPIKey: "test-api-key"}
	configLoaded = true
	configMutex.Unlock()

	// Test getting the API key
	apiKey, err := GetLunchMoneyAPIKey()
	if err != nil {
		t.Fatalf("Failed to get API key: %v", err)
	}

	// Verify the API key was retrieved correctly
	if apiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", apiKey)
	}

	// Test with empty API key
	configMutex.Lock()
	globalConfig = &Config{LunchMoneyAPIKey: ""}
	configMutex.Unlock()

	// Should return an error
	_, err = GetLunchMoneyAPIKey()
	if err == nil {
		t.Errorf("Expected error when API key is empty, got nil")
	}
}
