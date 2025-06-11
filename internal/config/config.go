package config

// Config represents the application configuration
type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	Server   ServerConfig   `mapstructure:"server"`
	LogLevel string         `mapstructure:"log_level"`
}

// TelegramConfig holds the Telegram bot configuration
type TelegramConfig struct {
	Token    string  `mapstructure:"token"`
	AdminIDs []int64 `mapstructure:"admin_ids"`
}

// ServerConfig holds the configuration for an X-ray server
type ServerConfig struct {
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	APIURL       string `mapstructure:"api_url"`
	SubURLPrefix string `mapstructure:"sub_url_prefix"`
}
