package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database      DatabaseConfig      `mapstructure:"database"`
	MasterService MasterServiceConfig `mapstructure:"master_service"`
}

// MasterServiceConfig points risk-service at the master-service that owns this
// stack's Location master data.
//
// Env overrides (via the "." -> "_" replacer): MASTER_SERVICE_URL,
// MASTER_SERVICE_TIMEOUT_SECONDS, MASTER_SERVICE_CACHE_TTL_SECONDS.
type MasterServiceConfig struct {
	URL             string `mapstructure:"url"`
	TimeoutSeconds  int    `mapstructure:"timeout_seconds"`
	CacheTTLSeconds int    `mapstructure:"cache_ttl_seconds"`
}

// defaultMasterServicePort is the port every master-service container listens
// on inside the docker network (shared and tenant stacks alike).
const defaultMasterServicePort = "8002"

// BaseURL returns the master-service base URL for this stack.
//
// An explicit url (MASTER_SERVICE_URL) always wins. Without one, the URL is
// derived per stack, never hard-wired to the shared service: tenant stacks
// share the rb_audit_network bridge with the control plane, so the bare alias
// "master-service" resolves to the *shared* master-service from inside a tenant
// container and would report another tenant's locations. Tenant services are
// named "<service>-<slug>" (see scripts/templates/docker-compose.tenant.yml.tpl)
// and set KAFKA_SERVICE_NAME to that name, so the tenant's own master-service is
// "master-service-<slug>".
func (c MasterServiceConfig) BaseURL() string {
	if u := strings.TrimSpace(c.URL); u != "" {
		return strings.TrimRight(u, "/")
	}
	return derivedMasterServiceURL(os.Getenv("KAFKA_SERVICE_NAME"))
}

func derivedMasterServiceURL(serviceName string) string {
	const prefix = "risk-service-"
	if slug := strings.TrimPrefix(strings.TrimSpace(serviceName), prefix); slug != serviceName && slug != "" {
		return "http://master-service-" + slug + ":" + defaultMasterServicePort
	}
	return "http://master-service:" + defaultMasterServicePort
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

func setDefaults() {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	// Registered as defaults so AutomaticEnv binds MASTER_SERVICE_* even though
	// config.yaml leaves them out. An empty url means "derive per stack".
	viper.SetDefault("master_service.url", "")
	viper.SetDefault("master_service.timeout_seconds", 2)
	viper.SetDefault("master_service.cache_ttl_seconds", 60)
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	// The replacer maps nested keys to conventional env var names, so
	// `database.name` is read from DATABASE_NAME. Without it viper looks up
	// the literal "DATABASE.NAME" and the override silently never applies —
	// which made per-tenant onboarding migrate the shared database instead.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
