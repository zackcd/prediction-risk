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

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
