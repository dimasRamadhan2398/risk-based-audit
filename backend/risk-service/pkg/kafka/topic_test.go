package kafka

import "testing"

func TestTopicName(t *testing.T) {
	cases := []struct{ prefix, want string }{
		{"", "audit-universe.events"},
		{"   ", "audit-universe.events"},
		{"bai", "bai.audit-universe.events"},
		{" bai. ", "bai.audit-universe.events"},
	}
	for _, c := range cases {
		if got := TopicName(c.prefix, BaseTopicAuditUniverse); got != c.want {
			t.Errorf("TopicName(%q) = %q, want %q", c.prefix, got, c.want)
		}
	}
}

// The prod stack sets no prefix and must keep publishing to the exact topic it
// always has; a tenant stack publishes to its own.
func TestAuditUniverseTopicFromEnv(t *testing.T) {
	t.Setenv(TopicPrefixEnv, "")
	if got := AuditUniverseTopicFromEnv(); got != "audit-universe.events" {
		t.Fatalf("empty prefix changed the topic: %q", got)
	}
	t.Setenv(TopicPrefixEnv, "bai")
	if got := AuditUniverseTopicFromEnv(); got != "bai.audit-universe.events" {
		t.Fatalf("KAFKA_TOPIC_PREFIX=bai gave %q", got)
	}
}
