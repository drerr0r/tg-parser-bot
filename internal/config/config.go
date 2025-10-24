package config

import (
	"os"

	"your-project/internal/models"

	"gopkg.in/yaml.v2"
)

type ConfigManager struct {
	config *models.Config
}

func NewConfigManager(configPath string) (*ConfigManager, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &ConfigManager{config: &config}, nil
}

func (cm *ConfigManager) GetConfig() *models.Config {
	return cm.config
}
