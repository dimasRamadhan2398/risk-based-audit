package config

import (
	"fmt"

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

type Config struct {
	Server           ServerConfig        `mapstructure:"server"`
	Database         DatabaseConfig      `mapstructure:"database"`
	Redis            RedisConfig         `mapstructure:"redis"`
	JWT              JWTConfig           `mapstructure:"jwt"`
	Log              LogConfig           `mapstructure:"log"`
	Kafka            KafkaConfig         `mapstructure:"kafka"`
	KafkaConsumer    KafkaConsumerConfig `mapstructure:"kafka_consumer"`
	App              AppMeta             `mapstructure:"app"      json:"app"`
	SMTP             SMTPConfig          `mapstructure:"smtp"`
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

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
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
	viper.SetDefault("smtp.host", "sandbox.smtp.mailtrap.io")
	viper.SetDefault("smtp.port", 2525)
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	setDefaults()
	// Allow environment variables to override
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
