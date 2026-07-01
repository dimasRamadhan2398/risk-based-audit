package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig    `mapstructure:"server"`
	Upstream UpstreamConfig  `mapstructure:"upstream"`
	Auth     AuthConfig      `mapstructure:"auth"`
	Log      LogConfig       `mapstructure:"log"`
	CORS     CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug, release
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// UpstreamConfig holds upstream service configuration
type UpstreamConfig struct {
	Services []ServiceConfig `mapstructure:"services"`
}

// ServiceConfig represents a single upstream service
type ServiceConfig struct {
	Name       string            `mapstructure:"name"`
	URL        string            `mapstructure:"url"`
	Prefix     string            `mapstructure:"prefix"`
	Methods    []string          `mapstructure:"methods"`
	Headers    map[string]string `mapstructure:"headers"`
	Timeout    int               `mapstructure:"timeout"` // seconds
	RetryCount int               `mapstructure:"retry_count"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled  bool `mapstructure:"enabled"`
	Interval int  `mapstructure:"interval"` // seconds
	Timeout  int  `mapstructure:"timeout"` // seconds
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	JWTSecret    string `mapstructure:"jwt_secret"`
	SkipPaths    []string `mapstructure:"skip_paths"`
	TokenHeader  string `mapstructure:"token_header"`
	TokenPrefix  string `mapstructure:"token_prefix"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // seconds
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	RequestsPerMin int    `mapstructure:"requests_per_min"`
	BurstSize     int     `mapstructure:"burst_size"`
	CleanupInterval int   `mapstructure:"cleanup_interval"` // seconds
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, console
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.write_timeout", 60)

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// CORS defaults
	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"})
	viper.SetDefault("cors.expose_headers", []string{"Content-Length", "Content-Type"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 3600)

	// Rate limit defaults
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_min", 100)
	viper.SetDefault("rate_limit.burst_size", 20)
	viper.SetDefault("rate_limit.cleanup_interval", 60)

	// Auth defaults
	viper.SetDefault("auth.enabled", true)
	viper.SetDefault("auth.token_header", "Authorization")
	viper.SetDefault("auth.token_prefix", "Bearer")
	viper.SetDefault("auth.skip_paths", []string{"/health", "/healthz", "/ready", "/metrics"})
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	setDefaults()
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
