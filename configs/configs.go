package configs

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		Swagger Swagger
		Crypto  Crypto
	}

	App struct {
		Name    string `env:"APP_NAME" envDefault:"gophkeeper-client"`
		Version string `env:"APP_VERSION" envDefault:"dev"`
	}

	HTTP struct {
		Host string `env:"HTTP_HOST" envDefault:"127.0.0.1"`
		Port string `env:"HTTP_PORT" envDefault:"8080"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL" envDefault:"debug"`
	}

	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	JWT struct {
		Secret string `env:"JWT_SECRET" envDefault:"supersecret"`
	}

	Crypto struct {
		Key string `env:"CRYPTO_KEY" envDefault:"change-me"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	godotenv.Load("./.env")
	godotenv.Load("../../.env")
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	host := flag.String("host", cfg.HTTP.Host, "server host")
	port := flag.String("port", cfg.HTTP.Port, "server port")
	logLevel := flag.String("log", cfg.Log.Level, "log level")
	cryptoKey := flag.String("crypto-key", cfg.Crypto.Key, "encryption key")
	ver := flag.Bool("v", false, "print version and exit")
	flag.Parse()

	if *ver {
		fmt.Println(cfg.App.Name, cfg.App.Version)
		return nil, nil
	}

	cfg.HTTP.Host = *host
	cfg.HTTP.Port = *port
	cfg.Log.Level = *logLevel
	cfg.Crypto.Key = *cryptoKey

	return cfg, nil
}
