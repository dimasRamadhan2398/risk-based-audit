package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeConsumerConfig mirrors the shape of the shipped config.yaml: topics and
// group id are set in the file, so the env vars must override file values.
func writeConsumerConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	yaml := "jwt:\n  secret: x\nkafka_consumer:\n  enabled: true\n" +
		"  group_id: \"audit-service-annual-plan-sync\"\n" +
		"  topics:\n    - \"audit-universe.events\"\n    - \"department.events\"\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// No prefix: the shared prod stack keeps its exact topics and group.
func TestKafkaTopicPrefixUnsetKeepsSharedNames(t *testing.T) {
	path := writeConsumerConfig(t)
	t.Setenv(KafkaTopicPrefixEnv, "")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"audit-universe.events", "department.events"}
	if !reflect.DeepEqual(cfg.KafkaConsumer.Topics, want) {
		t.Fatalf("topics changed without a prefix: %v", cfg.KafkaConsumer.Topics)
	}
	if cfg.KafkaConsumer.GroupID != DefaultAnnualPlanSyncGroupID {
		t.Fatalf("group changed without a prefix: %q", cfg.KafkaConsumer.GroupID)
	}
}

// KAFKA_TOPIC_PREFIX (the key risk-service reads) renames the Audit Universe
// topic only, and derives a per-tenant group when none is given.
func TestKafkaTopicPrefixFromEnv(t *testing.T) {
	path := writeConsumerConfig(t)
	t.Setenv(KafkaTopicPrefixEnv, "bai")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"bai.audit-universe.events", "department.events"}
	if !reflect.DeepEqual(cfg.KafkaConsumer.Topics, want) {
		t.Fatalf("KAFKA_TOPIC_PREFIX not applied: %v", cfg.KafkaConsumer.Topics)
	}
	if cfg.KafkaConsumer.GroupID != "audit-service-annual-plan-sync-bai" {
		t.Fatalf("group not derived from prefix: %q", cfg.KafkaConsumer.GroupID)
	}
	if got := TopicName("bai", "audit-universe.events"); got != want[0] {
		t.Fatalf("TopicName disagrees with Load: %q", got)
	}
}

// An explicit KAFKA_CONSUMER_GROUP_ID wins over both the file and the derived
// name.
func TestKafkaConsumerGroupIDFromEnv(t *testing.T) {
	path := writeConsumerConfig(t)
	t.Setenv(KafkaTopicPrefixEnv, "bai")
	t.Setenv("KAFKA_CONSUMER_GROUP_ID", "custom-group-bai")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.KafkaConsumer.GroupID != "custom-group-bai" {
		t.Fatalf("KAFKA_CONSUMER_GROUP_ID not applied: %q", cfg.KafkaConsumer.GroupID)
	}
}
