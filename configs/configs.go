package configs

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	// Config -.
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		Swagger Swagger
		Crypto  Crypto
	}

	// App -.
	App struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP -.
	HTTP struct {
		Host string `env:HTTP_HOST,required`
		Port string `env:"HTTP_PORT,required"`
	}

	// Log -.
	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// Swagger -.
	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	// JWT -.
	JWT struct {
		Secret string `env:"JWT_SECRET" envDefault:"supersecret"`
	}

	// Crypto -.
	Crypto struct {
		Key string `env:"CRYPTO_KEY,required"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	godotenv.Load("./.env")
	godotenv.Load("../../.env")
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
