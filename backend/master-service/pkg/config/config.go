package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type KafkaConfig struct {
	Enabled    bool     `mapstructure:"enabled"`
	Brokers    []string `mapstructure:"brokers"`
	Topic      string   `mapstructure:"topic"`
	ServiceName string  `mapstructure:"service_name"`
}

type KafkaConsumerConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Brokers  []string `mapstructure:"brokers"`
	GroupID  string   `mapstructure:"group_id"`
	Topics   []string `mapstructure:"topics"`
	Version  string   `mapstructure:"version"`
	Assignor string   `mapstructure:"assignor"` // range, roundrobin, sticky
}

// AuthServiceConfig locates auth-service. Employee creation provisions login
// accounts there, so URL must be set for @aifl.co.id employees (AUTH_SERVICE_URL).
type AuthServiceConfig struct {
	URL            string `mapstructure:"url"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

// OrgStructureEnv is the opt-in flag for the flexible organisation structure
// (docs/org-structure-model-plan.md section 3.2). It is written only into a NEW
// tenant's compose file. Unset (the default) means legacy: flat departments and
// no org-structure DDL at all.
const OrgStructureEnv = "ORG_STRUCTURE_MODE"

// OrgStructureConfig carries the raw flag. Use FlexibleRequested to read it;
// the effective mode also needs the org_levels table (pkg/orgmode).
type OrgStructureConfig struct {
	Mode string `mapstructure:"mode"`
}

// FlexibleRequested reports whether ORG_STRUCTURE_MODE asks for the flexible
// structure. Only the exact word "flexible" (any case) turns it on.
func (c OrgStructureConfig) FlexibleRequested() bool {
	return strings.EqualFold(strings.TrimSpace(c.Mode), "flexible")
}

type Config struct {
	Server           ServerConfig        `mapstructure:"server"`
	Database         DatabaseConfig      `mapstructure:"database"`
	Redis            RedisConfig         `mapstructure:"redis"`
	JWT              JWTConfig           `mapstructure:"jwt"`
	Log              LogConfig           `mapstructure:"log"`
	Kafka            KafkaConfig         `mapstructure:"kafka"`
	KafkaConsumer    KafkaConsumerConfig `mapstructure:"kafka_consumer"`
	AuthService      AuthServiceConfig  `mapstructure:"auth_service"`
	OrgStructure     OrgStructureConfig  `mapstructure:"org_structure"`
	App              AppMeta             `mapstructure:"app"      json:"app"`
}

type AppMeta struct {
	AppName               string   `mapstructure:"app_name" json:"app_name"`
	AppEnv                string   `mapstructure:"app_env" json:"app_env"`
	SignatureKey          string   `mapstructure:"signature_key" json:"signature_key"`
	RateLimiterMax  float64 `mapstructure:"rate_limiter_max"  json:"rate_limiter_max"`
	RateLimiterTime int     `mapstructure:"rate_limiter_time" json:"rate_limiter_time"`
}
type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Mode        string `mapstructure:"mode"` // debug, release
	ReadTimeout int    `mapstructure:"read_timeout"`
}

type DatabaseConfig struct {
	Host                  string `mapstructure:"host"                    json:"host"`
	Port                  int    `mapstructure:"port"                    json:"port"`
	Name                  string `mapstructure:"name"                    json:"name"`
	Username              string `mapstructure:"username"                json:"username"`
	Password              string `mapstructure:"password"                json:"password"`
	SSLMode               string `mapstructure:"sslmode"                 json:"sslmode"`
	MaxOpenConnections    int    `mapstructure:"max_open_connections"    json:"max_open_connections"`
	MaxLifeTimeConnection int    `mapstructure:"max_lifetime_connection" json:"max_lifetime_connection"`
	MaxIdleConnections    int    `mapstructure:"max_idle_connections"    json:"max_idle_connections"`
	MaxIdleTime           int    `mapstructure:"max_idle_time"           json:"max_idle_time"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiryHour int    `mapstructure:"expiry_hour"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json, console
}

func setDefaults() {
	viper.SetDefault("app.appName", "Risk Based Internal Audit Auth Service")
	viper.SetDefault("app.appEnv", "development")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.max_open_connections", 25)
	viper.SetDefault("database.max_lifetime_connection", 300)
	viper.SetDefault("database.max_idle_connections", 10)
	viper.SetDefault("database.max_idle_time", 60)
	viper.SetDefault("database.sslmode", "require")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.expiry_hour", 24)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("kafka.enabled", false)
	viper.SetDefault("kafka.topic", "service-logs")
	viper.SetDefault("kafka.service_name", "auth-service")
	viper.SetDefault("kafka_consumer.enabled", false)
	viper.SetDefault("kafka_consumer.group_id", "auth-service-consumer")
	viper.SetDefault("kafka_consumer.topics", []string{"events"})
	viper.SetDefault("kafka_consumer.version", "3.6.0")
	viper.SetDefault("kafka_consumer.assignor", "roundrobin")
	viper.SetDefault("auth_service.url", "")
	viper.SetDefault("auth_service.timeout_seconds", 10)
	viper.SetDefault("org_structure.mode", "")
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	setDefaults()
	// Allow environment variables to override.
	// The replacer maps nested keys to conventional env var names, so
	// `database.name` is read from DATABASE_NAME. Without it viper looks up
	// the literal "DATABASE.NAME" and the override silently never applies —
	// which made per-tenant onboarding migrate the shared database instead.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	// jwt.secret is bound explicitly so JWT_SECRET overrides the file even when the file has no secret key.
	_ = viper.BindEnv("jwt.secret", JWTSecretEnv)
	// Bound explicitly for the same reason: config.yaml has no org_structure
	// section, so AutomaticEnv alone would never see the variable on Unmarshal.
	_ = viper.BindEnv("org_structure.mode", OrgStructureEnv)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
