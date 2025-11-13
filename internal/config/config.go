package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	DatabaseURL    string
	Port           string
	DataDir        string
	TrustedOrigins []string
}

// Load loads configuration from multiple sources with priority:
// 1. Command flags (set via viper.Set)
// 2. Config file (~/.kaunta/config.toml or ./kaunta.toml)
// 3. Environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName("kaunta")
	v.SetConfigType("toml")

	// Add config search paths
	// 1. Current directory
	v.AddConfigPath(".")

	// 2. User home directory (~/.kaunta/)
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".kaunta"))
	}

	// Set default values
	v.SetDefault("port", "3000")
	v.SetDefault("data_dir", "./data")
	v.SetDefault("trusted_origins", "localhost")

	// Bind environment variables
	v.SetEnvPrefix("") // No prefix, allow DATABASE_URL directly
	v.AutomaticEnv()

	// Map environment variable names to config keys
	_ = v.BindEnv("database_url", "DATABASE_URL")
	_ = v.BindEnv("port", "PORT")
	_ = v.BindEnv("data_dir", "DATA_DIR")
	_ = v.BindEnv("trusted_origins", "TRUSTED_ORIGINS")

	// Read config file if it exists (don't error if not found)
	_ = v.ReadInConfig()

	// Parse comma-separated trusted origins
	originsStr := v.GetString("trusted_origins")
	origins := parseTrustedOrigins(originsStr)

	return &Config{
		DatabaseURL:    v.GetString("database_url"),
		Port:           v.GetString("port"),
		DataDir:        v.GetString("data_dir"),
		TrustedOrigins: origins,
	}, nil
}

// LoadWithOverrides loads config and applies flag overrides
func LoadWithOverrides(databaseURL, port, dataDir string) (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName("kaunta")
	v.SetConfigType("toml")

	// Add config search paths
	v.AddConfigPath(".")
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".kaunta"))
	}

	// Set default values
	v.SetDefault("port", "3000")
	v.SetDefault("data_dir", "./data")
	v.SetDefault("trusted_origins", "localhost")

	// Bind environment variables
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	_ = v.BindEnv("database_url", "DATABASE_URL")
	_ = v.BindEnv("port", "PORT")
	_ = v.BindEnv("data_dir", "DATA_DIR")
	_ = v.BindEnv("trusted_origins", "TRUSTED_ORIGINS")

	// Read config file
	_ = v.ReadInConfig()

	// Apply flag overrides (highest priority)
	if databaseURL != "" {
		v.Set("database_url", databaseURL)
	}
	if port != "" {
		v.Set("port", port)
	}
	if dataDir != "" {
		v.Set("data_dir", dataDir)
	}

	// Parse comma-separated trusted origins
	originsStr := v.GetString("trusted_origins")
	origins := parseTrustedOrigins(originsStr)

	return &Config{
		DatabaseURL:    v.GetString("database_url"),
		Port:           v.GetString("port"),
		DataDir:        v.GetString("data_dir"),
		TrustedOrigins: origins,
	}, nil
}

// parseTrustedOrigins parses a comma-separated string into a slice of trimmed, lowercased origins
func parseTrustedOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{}
	}

	parts := strings.Split(originsStr, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		origin = strings.ToLower(origin)
		// Strip protocol if provided
		origin = strings.TrimPrefix(origin, "http://")
		origin = strings.TrimPrefix(origin, "https://")
		origin = strings.TrimSuffix(origin, "/")

		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}
