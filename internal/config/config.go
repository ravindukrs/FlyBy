package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Token represents the authentication token structure
type Token struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

// Target represents a Concourse target configuration
type Target struct {
	Name       string `yaml:"name"`
	API        string `yaml:"api"`        // fly CLI uses 'api' not 'url'
	Team       string `yaml:"team"`
	Token      *Token `yaml:"token,omitempty"` // Token is a nested object
	Insecure   bool   `yaml:"insecure,omitempty"`
	CACert     string `yaml:"ca_cert,omitempty"`
	ClientCert string `yaml:"client_cert,omitempty"`
	ClientKey  string `yaml:"client_key,omitempty"`
}

// GetURL returns the API URL for compatibility
func (t Target) GetURL() string {
	return t.API
}

// GetTokenValue returns the token value if available
func (t Target) GetTokenValue() string {
	if t.Token != nil {
		return t.Token.Value
	}
	return ""
}

// HasToken returns true if the target has a valid token
func (t Target) HasToken() bool {
	return t.Token != nil && t.Token.Value != ""
}

// FlyConfig represents the ~/.flyrc configuration
type FlyConfig struct {
	Targets map[string]Target `yaml:"targets"`
}

// ConfigManager handles fly configuration operations
type ConfigManager struct {
	configPath string
	config     *FlyConfig
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".flyrc")
	manager := &ConfigManager{
		configPath: configPath,
		config:     &FlyConfig{Targets: make(map[string]Target)},
	}

	if err := manager.LoadConfig(); err != nil {
		// If config doesn't exist, start with empty config
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return manager, nil
}

// LoadConfig loads the fly configuration from ~/.flyrc
func (cm *ConfigManager) LoadConfig() error {
	data, err := ioutil.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cm.config)
}

// SaveConfig saves the configuration to ~/.flyrc
func (cm *ConfigManager) SaveConfig() error {
	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return ioutil.WriteFile(cm.configPath, data, 0600)
}

// GetTargets returns all configured targets
func (cm *ConfigManager) GetTargets() map[string]Target {
	return cm.config.Targets
}

// GetTarget returns a specific target by name
func (cm *ConfigManager) GetTarget(name string) (Target, bool) {
	target, exists := cm.config.Targets[name]
	return target, exists
}

// AddTarget adds a new target to the configuration
func (cm *ConfigManager) AddTarget(name, url, team string) error {
	if name == "" || url == "" || team == "" {
		return fmt.Errorf("name, url, and team are required")
	}

	cm.config.Targets[name] = Target{
		Name: name,
		API:  url,  // Use API field instead of URL
		Team: team,
	}

	return cm.SaveConfig()
}

// RemoveTarget removes a target from the configuration
func (cm *ConfigManager) RemoveTarget(name string) error {
	if _, exists := cm.config.Targets[name]; !exists {
		return fmt.Errorf("target '%s' does not exist", name)
	}

	delete(cm.config.Targets, name)
	return cm.SaveConfig()
}

// UpdateTarget updates an existing target
func (cm *ConfigManager) UpdateTarget(name string, target Target) error {
	if _, exists := cm.config.Targets[name]; !exists {
		return fmt.Errorf("target '%s' does not exist", name)
	}

	target.Name = name
	cm.config.Targets[name] = target
	return cm.SaveConfig()
}

// GetConfigPath returns the path to the fly config file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// ListTeams returns a list of unique teams from all targets
func (cm *ConfigManager) ListTeams() []string {
	teams := make(map[string]bool)
	var result []string

	for _, target := range cm.config.Targets {
		if target.Team != "" && !teams[target.Team] {
			teams[target.Team] = true
			result = append(result, target.Team)
		}
	}

	return result
}

// GetTargetsByTeam returns all targets for a specific team
func (cm *ConfigManager) GetTargetsByTeam(team string) []Target {
	var targets []Target

	for _, target := range cm.config.Targets {
		if target.Team == team {
			targets = append(targets, target)
		}
	}

	return targets
}

// ReadFlyrcRaw reads and returns the raw content of ~/.flyrc file
// This is useful for debugging or showing the raw config to users
func (cm *ConfigManager) ReadFlyrcRaw() (string, error) {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("flyrc file does not exist at %s", cm.configPath)
	}

	file, err := os.Open(cm.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to open flyrc file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read flyrc file: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}