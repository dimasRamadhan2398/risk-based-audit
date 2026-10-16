package config

import (
	"os"
	"path/filepath"
	"testing"
)

// The annual plan consumer reads kafka_consumer.* only. KAFKA_BROKERS (the
// service-logs producer) must not change its brokers, and KAFKA_CONSUMER_*
// must override the file.
func TestKafkaConsumerEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := "jwt:\n  secret: x\nkafka:\n  brokers:\n    - \"kafka:9092\"\nkafka_consumer:\n  enabled: true\n  brokers:\n    - \"kafka:29092\"\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KAFKA_BROKERS", "logs-broker:9092")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.KafkaConsumer.Brokers) != 1 || cfg.KafkaConsumer.Brokers[0] != "kafka:29092" {
		t.Fatalf("KAFKA_BROKERS changed the consumer brokers: %v", cfg.KafkaConsumer.Brokers)
	}
	if len(cfg.Kafka.Brokers) != 1 || cfg.Kafka.Brokers[0] != "logs-broker:9092" {
		t.Fatalf("KAFKA_BROKERS did not set the producer brokers: %v", cfg.Kafka.Brokers)
	}

	t.Setenv("KAFKA_CONSUMER_BROKERS", "b1:29092,b2:29092")
	t.Setenv("KAFKA_CONSUMER_ENABLED", "false")
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.KafkaConsumer.Brokers) != 2 {
		t.Fatalf("KAFKA_CONSUMER_BROKERS not applied: %v", cfg.KafkaConsumer.Brokers)
	}
	if cfg.KafkaConsumer.Enabled {
		t.Fatal("KAFKA_CONSUMER_ENABLED=false not applied")
	}
}
