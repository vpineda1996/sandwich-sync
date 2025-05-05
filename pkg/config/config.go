package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type RogersOptions struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DeviceId string `yaml:"deviceId"`
}

type WealthsimpleOptions struct {
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	PrevSession string `yaml:"prevSession"`
}

// Config holds the application configuration
type Config struct {
	LunchMoneyAPIKey       string              `yaml:"lunchMoneyApiKey"`
	RogersApiOptions       RogersOptions       `yaml:"rogers"`
	WealthsimpleApiOptions WealthsimpleOptions `yaml:"wealthsimple"`
}

var (
	// Global configuration instance
	globalConfig *Config
	// Mutex to ensure thread-safe access to the global configuration
	configMutex sync.RWMutex
	// Flag to track if the configuration has been loaded
	configLoaded bool
)

// LoadConfig loads the configuration from the specified YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the YAML data
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// InitGlobalConfig initializes the global configuration from the specified file
func InitGlobalConfig(configPath string) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig = config
	configLoaded = true
	return nil
}

// GetConfig returns the global configuration instance
// If the configuration hasn't been loaded yet, it attempts to load it from
// the default location (./config.yaml)
func GetConfig() (*Config, error) {
	configMutex.RLock()
	if configLoaded {
		defer configMutex.RUnlock()
		return globalConfig, nil
	}
	configMutex.RUnlock()

	// Try to load from default location
	configPath := "config.yaml"
	if err := InitGlobalConfig(configPath); err != nil {
		// If the default config file doesn't exist, create it
		if os.IsNotExist(err) {
			// Ensure the directory exists
			dir := filepath.Dir(configPath)
			if dir != "" && dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return nil, fmt.Errorf("error creating config directory: %w", err)
				}
			}

			// Create a default configuration
			defaultConfig := &Config{
				LunchMoneyAPIKey: "", // Empty by default
				RogersApiOptions: RogersOptions{
					Username: "",
					Password: "",
					DeviceId: "",
				},
			}

			// Marshal the default configuration to YAML
			data, err := yaml.Marshal(defaultConfig)
			if err != nil {
				return nil, fmt.Errorf("error creating default config: %w", err)
			}

			// Write the default configuration to the file
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				return nil, fmt.Errorf("error writing default config: %w", err)
			}

			// Set the global configuration to the default
			configMutex.Lock()
			globalConfig = defaultConfig
			configLoaded = true
			configMutex.Unlock()

			return defaultConfig, nil
		}
		return nil, err
	}

	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig, nil
}

// GetRogersCredentials returns the Rogers API credentials from the configuration
func GetRogersCredentials() (string, string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", "", err
	}

	if config.RogersApiOptions.Username == "" || config.RogersApiOptions.Password == "" {
		return "", "", fmt.Errorf("error: Rogers API credentials not set in configuration")
	}

	return config.RogersApiOptions.Username, config.RogersApiOptions.Password, nil
}

func GetRogersDeviceId() (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	if config.RogersApiOptions.DeviceId == "" {
		return "", fmt.Errorf("error: Rogers API device fingerprint not set in configuration")
	}

	return config.RogersApiOptions.DeviceId, nil
}

// GetLunchMoneyAPIKey returns the Lunch Money API key from the configuration
func GetLunchMoneyAPIKey() (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	if config.LunchMoneyAPIKey == "" {
		return "", fmt.Errorf("lunch money API key not set in configuration")
	}

	return config.LunchMoneyAPIKey, nil
}

// GetWealthsimpleCredentials returns the Wealthsimple API credentials from the configuration
func GetWealthsimpleCredentials() (string, string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", "", err
	}

	if config.WealthsimpleApiOptions.Username == "" || config.WealthsimpleApiOptions.Password == "" {
		return "", "", fmt.Errorf("error: Wealthsimple API credentials not set in configuration")
	}

	return config.WealthsimpleApiOptions.Username, config.WealthsimpleApiOptions.Password, nil
}

func GetWealthsimplePrevSession() (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	if config.WealthsimpleApiOptions.PrevSession == "" {
		return "", fmt.Errorf("error: Wealthsimple API prev session not set in configuration")
	}

	return config.WealthsimpleApiOptions.PrevSession, nil
}

func SetWealthsimplePrevSession(session string) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	config.WealthsimpleApiOptions.PrevSession = session

	configMutex.Lock()
	defer configMutex.Unlock()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	if err := os.WriteFile("config.yaml", data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
