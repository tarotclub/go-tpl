package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all application configuration values.
type Config struct {
	App  AppConfig  `mapstructure:"app"`
	Log  LogConfig  `mapstructure:"log"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// Load reads configuration from the given file path.
// The file must be a YAML file named config.yaml (or the path provided).
func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	// Allow environment variables to override config values.
	viper.AutomaticEnv()

	// Set sensible defaults.
	viper.SetDefault("app.name", "go-tpl")
	viper.SetDefault("app.version", "dev")
	viper.SetDefault("log.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}
