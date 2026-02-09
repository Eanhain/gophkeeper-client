// Package configs handles application configuration.
//
// Configuration values are resolved in the following priority order
// (highest wins):
//  1. CLI flags (--host, --port, --log, --crypto-key)
//  2. Environment variables (HTTP_HOST, HTTP_PORT, …)
//  3. .env files (./.env, ../../.env)
//  4. Built-in defaults (envDefault tags)
package configs

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	// Config is the root configuration container.
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		Swagger Swagger
		Crypto  Crypto
	}

	// App describes the application metadata (name, version).
	App struct {
		Name    string `env:"APP_NAME" envDefault:"gophkeeper-client"`
		Version string `env:"APP_VERSION" envDefault:"dev"`
	}

	// HTTP holds the server address to which the client connects.
	HTTP struct {
		Host string `env:"HTTP_HOST" envDefault:"127.0.0.1"`
		Port string `env:"HTTP_PORT" envDefault:"8080"`
	}

	// Log configures the log level (debug, info, warn, error).
	Log struct {
		Level string `env:"LOG_LEVEL" envDefault:"debug"`
	}

	// Swagger is not used by the client; kept for config parity with the server.
	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	// JWT is not used by the client directly; the token is received from the server.
	JWT struct {
		Secret string `env:"JWT_SECRET" envDefault:"supersecret"`
	}

	// Crypto holds the symmetric key used both for local cache encryption
	// and for encrypting HTTP bodies sent to/from the server.
	Crypto struct {
		Key string `env:"CRYPTO_KEY" envDefault:"change-me"`
	}
)

// NewConfig loads the configuration and parses CLI flags.
// Returns (nil, nil) when the -v flag is used (version is printed to stdout).
func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Try loading .env files — errors are silently ignored (files are optional).
	godotenv.Load("./.env")
	godotenv.Load("../../.env")
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// CLI flags override env/defaults.
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
