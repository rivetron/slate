package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rivetron/slate/src/models"
	"gopkg.in/yaml.v3"
)

const (
	configFileName = "slate.yaml"
)

type Loader struct {
	configPath string
}

func New() *Loader {
	return &Loader{}
}

func ExampleConfig() string {
	config := models.NewDefaultConfig()
	data, err := yaml.Marshal(config)
	if err != nil {
		return "Failed to fetch example config"
	}

	return string(data)
}

func CreateDefaultConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "slate")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, configFileName)

	// ? Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return "", fmt.Errorf("config file already exists at %s", configPath)
	}

	// * Create default config
	config := models.NewDefaultConfig()

	// * Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshall config: %w", err)
	}

	// * Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

func ValidateConfig(config *models.Config) error {
	// * Validate theme mode
	if config.Theme.Mode != "" {
		validModes := map[string]bool{"auto": true, "dark": true, "light": true}
		if !validModes[config.Theme.Mode] {
			return fmt.Errorf("invalid theme mode: %s (must be auto, dark, or light)", config.Theme.Mode)
		}
	}

	// * Validate presentation settings
	if config.Presentation.WordWrap < 0 {
		return fmt.Errorf("word wrap must be non-negative")
	}

	if config.Presentation.Margin < 0 {
		return fmt.Errorf("margin must be non-negative")
	}

	if config.Presentation.Padding < 0 {
		return fmt.Errorf("padding must be non-negative")
	}

	// * Validate keybindings
	if len(config.Keybindings.Next) == 0 {
		return fmt.Errorf("next keybinding must have at least one key")
	}

	if len(config.Keybindings.Previous) == 0 {
		return fmt.Errorf("previous keybinding must have at least one key")
	}

	if len(config.Keybindings.Quit) == 0 {
		return fmt.Errorf("quit keybinding must have at least one key")
	}

	return nil
}

func (l *Loader) getSearchPaths() []string {
	paths := make([]string, 0)

	// * Current directory
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, configFileName))
		paths = append(paths, filepath.Join(cwd, "."+configFileName))
	}

	// * Home directory config folder
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".config", "slate", configFileName))
		paths = append(paths, filepath.Join(homeDir, "."+configFileName))
	}

	// * XDG_CONFIG_HOME
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		paths = append(paths, filepath.Join(xdgConfig, "slate", configFileName))
	}

	return paths
}

func (l *Loader) GetConfigPath() string {
	return l.configPath
}

func (l *Loader) FindConfig() (string, error) {
	searchPaths := l.getSearchPaths()

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no config file found in search pattern")
}

func (l *Loader) Save(config *models.Config) error {
	configPath := l.configPath

	// * Determine path
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configDir := filepath.Join(homeDir, ".config", "slate")
		if err := os.MkdirAll(configDir, 0750); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configPath = filepath.Join(configDir, configFileName)
		l.configPath = configPath
	}

	// * Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// * Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (l *Loader) LoadFromFile(path string) (*models.Config, error) {
	// * Clean and validate the path
	cleanPath := filepath.Clean(path)

	// * Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	// ? Check if file exists and is a regular file (not a directory or device)
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access config file: %w", err)
	}

	if !info.Mode().IsRegular() {
		return nil, errors.New("config path must be a regular file")
	}

	// #nosec G304 -- path is cleaned, validated, and comes from CLI argument
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return &config, nil
}
func (l *Loader) Load() (*models.Config, error) {
	// Start with default config
	config := models.NewDefaultConfig()

	// Try to find and load config file
	configPath, err := l.FindConfig()
	if err != nil {
		// No config file found, use defaults
		return config, nil
	}

	l.configPath = configPath

	// Load config from file
	fileConfig, err := l.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	// Merge file config into default config
	config.Merge(fileConfig)

	return config, nil
}
