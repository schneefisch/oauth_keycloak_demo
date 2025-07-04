package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string `koanf:"port"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     string `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Name     string `koanf:"name"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	KeycloakURL   string `koanf:"keycloak_url"`
	ClientID      string `koanf:"client_id"`
	ClientSecret  string `koanf:"client_secret"`
	RequiredScope string `koanf:"required_scope"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "postgres",
			Port:     "5432",
			User:     "events_user",
			Password: "events_password",
			Name:     "events_demo",
		},
		Auth: AuthConfig{
			KeycloakURL:   "http://localhost:8081",
			ClientID:      "events-api",
			RequiredScope: "events-api-access",
		},
	}
}

// Load loads configuration from environment variables and optional config file
func Load(configFile string) (*Config, error) {
	k := koanf.New(".")

	// Load default config
	cfg := DefaultConfig()

	// Load from .env file if it exists (useful for local development)
	if configFile != "" {
		if err := k.Load(file.Provider(configFile), dotenv.Parser()); err != nil {
			log.Printf("Warning: error loading config file: %v", err)
		}
	}

	// Load from environment variables (overrides defaults and file config)
	// Use prefix "" and delimiter "_" to load all environment variables
	if err := k.Load(env.Provider("", "_", func(s string) string {
		// Convert environment variables to lowercase and replace _ with .
		// e.g., SERVER_PORT becomes server.port
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	}), nil); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	// Unmarshal into Config struct
	if err := k.Unmarshal("server", &cfg.Server); err != nil {
		return nil, fmt.Errorf("error unmarshaling server config: %w", err)
	}
	if err := k.Unmarshal("db", &cfg.Database); err != nil {
		return nil, fmt.Errorf("error unmarshaling database config: %w", err)
	}
	if err := k.Unmarshal("", &cfg.Auth); err != nil {
		return nil, fmt.Errorf("error unmarshaling auth config: %w", err)
	}

	// Special handling for CLIENT_SECRET environment variable
	// This is needed because the environment variable name doesn't match the expected format
	if clientSecret := os.Getenv("CLIENT_SECRET"); clientSecret != "" {
		cfg.Auth.ClientSecret = clientSecret
	}

	// Special handling for KEYCLOAK_URL environment variable
	// This is needed to ensure the correct URL is used in Docker environment
	if keycloakURL := os.Getenv("KEYCLOAK_URL"); keycloakURL != "" {
		cfg.Auth.KeycloakURL = keycloakURL
	}

	// Special handling for REQUIRED_SCOPE environment variable
	// This is needed to ensure the correct scope is used for token validation
	if requiredScope := os.Getenv("REQUIRED_SCOPE"); requiredScope != "" {
		cfg.Auth.RequiredScope = requiredScope
	}

	return cfg, nil
}

// TestConfig creates a configuration for testing with the given overrides
func TestConfig(overrides *Config) *Config {
	cfg := DefaultConfig()

	// Apply overrides if provided
	if overrides != nil {
		// Server overrides
		if overrides.Server.Port != "" {
			cfg.Server.Port = overrides.Server.Port
		}

		// Database overrides
		if overrides.Database.Host != "" {
			cfg.Database.Host = overrides.Database.Host
		}
		if overrides.Database.Port != "" {
			cfg.Database.Port = overrides.Database.Port
		}
		if overrides.Database.User != "" {
			cfg.Database.User = overrides.Database.User
		}
		if overrides.Database.Password != "" {
			cfg.Database.Password = overrides.Database.Password
		}
		if overrides.Database.Name != "" {
			cfg.Database.Name = overrides.Database.Name
		}

		// Auth overrides
		if overrides.Auth.KeycloakURL != "" {
			cfg.Auth.KeycloakURL = overrides.Auth.KeycloakURL
		}
		if overrides.Auth.ClientID != "" {
			cfg.Auth.ClientID = overrides.Auth.ClientID
		}
		if overrides.Auth.ClientSecret != "" {
			cfg.Auth.ClientSecret = overrides.Auth.ClientSecret
		}
		if overrides.Auth.RequiredScope != "" {
			cfg.Auth.RequiredScope = overrides.Auth.RequiredScope
		}
	}

	return cfg
}

// ConnectionString returns a PostgreSQL connection string based on the database configuration
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name)
}
