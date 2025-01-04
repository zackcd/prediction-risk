package config

import "github.com/spf13/viper"

type Config struct {
	Environment string
	Server      struct {
		Port int
		Host string
	}
	Kalshi struct {
		BaseURL    string
		APIKeyID   string
		PrivateKey string
	}
	IsDryRun  bool
	Databases struct {
		User     string
		Password string
		Name     string
		Port     int
		Host     string
	}
}

func LoadConfig() (*Config, error) {
	viper.SetEnvPrefix("") // No prefix for env vars
	viper.AutomaticEnv()   // Automatically read env vars

	// Map the env vars to config keys
	viper.BindEnv("Environment", "APP_ENV")
	viper.BindEnv("Server.Port", "SERVER_PORT")
	viper.BindEnv("Server.Host", "SERVER_HOST")
	viper.BindEnv("Kalshi.BaseURL", "KALSHI_BASE_URL")
	viper.BindEnv("Kalshi.APIKeyID", "KALSHI_API_KEY")
	viper.BindEnv("Kalshi.PrivateKey", "KALSHI_PRIVATE_KEY")
	viper.SetDefault("isDryRun", true)
	viper.BindEnv("isDryRun", "IS_DRY_RUN")
	viper.BindEnv("Databases.User", "DB_USER")
	viper.BindEnv("Databases.Password", "DB_PASSWORD")
	viper.BindEnv("Databases.Name", "DB_NAME")
	viper.BindEnv("Databases.Port", "DB_PORT")
	viper.BindEnv("Databases.Host", "DB_HOST")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
