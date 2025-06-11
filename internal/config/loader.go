package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Set default values
	v.SetDefault("log_level", "info")

	// Define environment variables
	v.BindEnv("TG_TOKEN")
	v.BindEnv("TG_ADMIN_IDS")
	v.BindEnv("XRAY_USER")
	v.BindEnv("XRAY_PASSWORD")
	v.BindEnv("XRAY_API_URL")
	v.BindEnv("XRAY_SUB_URL_PREFIX")

	// Create config instance
	cfg := &Config{
		LogLevel: v.GetString("log_level"),
		Telegram: TelegramConfig{
			Token: v.GetString("TG_TOKEN"),
		},
	}

	// Parse admin IDs
	adminIDsStr := v.GetString("TG_ADMIN_IDS")
	if adminIDsStr != "" {
		adminIDsSlice := strings.Split(adminIDsStr, ",")
		adminIDs := make([]int64, 0, len(adminIDsSlice))
		for _, idStr := range adminIDsSlice {
			var id int64
			if _, err := fmt.Sscanf(strings.TrimSpace(idStr), "%d", &id); err == nil {
				adminIDs = append(adminIDs, id)
			}
		}
		cfg.Telegram.AdminIDs = adminIDs
	}

	// Parse server configuration
	user := v.GetString("XRAY_USER")
	password := v.GetString("XRAY_PASSWORD")
	apiURL := v.GetString("XRAY_API_URL")
	subURLPrefix := v.GetString("XRAY_SUB_URL_PREFIX")

	if user == "" || password == "" || apiURL == "" {
		return nil, errors.New("missing required server configuration")
	}

	// Create server configuration
	cfg.Server = ServerConfig{
		User:         strings.TrimSpace(user),
		Password:     strings.TrimSpace(password),
		APIURL:       strings.TrimSpace(apiURL),
		SubURLPrefix: strings.TrimSpace(subURLPrefix),
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg.Telegram.Token == "" {
		return errors.New("TG_TOKEN is required")
	}

	if len(cfg.Telegram.AdminIDs) == 0 {
		return errors.New("TG_ADMIN_IDS is required")
	}

	// Validate server configuration
	if cfg.Server.User == "" {
		return errors.New("server user is required")
	}
	if cfg.Server.Password == "" {
		return errors.New("server password is required")
	}
	if cfg.Server.APIURL == "" {
		return errors.New("server API URL is required")
	}

	return nil
}
