package activitylog

import (
	"os"
	"strings"
)

// Env vars read by this package (see config.ActivityLogConfig for the rest).
const (
	// TopicPrefixEnv is the same key risk-service and audit-service read.
	TopicPrefixEnv = "KAFKA_TOPIC_PREFIX"
	// ServiceNameEnv names this stack's auth-service ("auth-service-<slug>" on
	// a tenant stack, "auth-service" on the shared one).
	ServiceNameEnv = "KAFKA_SERVICE_NAME"
)

// BaseTopic is the unprefixed activity topic.
const BaseTopic = "activity-logs"

// DefaultServiceName is the shared (control-plane) auth-service name.
const DefaultServiceName = "auth-service"

// TopicName returns base, or "<prefix>.<base>" when prefix is non-empty
// (same rule as risk-service pkg/kafka.TopicName).
func TopicName(prefix, base string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), ".")
	if prefix == "" {
		return base
	}
	return prefix + "." + base
}

// TenantSlug returns the tenant slug encoded in an "auth-service-<slug>"
// service name, or "" for the shared stack.
func TenantSlug(serviceName string) string {
	serviceName = strings.TrimSpace(serviceName)
	if slug := strings.TrimPrefix(serviceName, DefaultServiceName+"-"); slug != serviceName {
		return strings.TrimSpace(slug)
	}
	return ""
}

// ResolveTopic picks this stack's activity topic. An explicit topic wins;
// otherwise the prefix is KAFKA_TOPIC_PREFIX, falling back to the tenant slug
// in the service name, because tenant auth-services do not set
// KAFKA_TOPIC_PREFIX today. Either way each tenant gets its own topic.
func ResolveTopic(explicit, prefix, serviceName string) string {
	if t := strings.TrimSpace(explicit); t != "" {
		return t
	}
	if strings.TrimSpace(prefix) == "" {
		prefix = TenantSlug(serviceName)
	}
	return TopicName(prefix, BaseTopic)
}

// ResolveSource is the identity stamped on every Entry and the only source
// this stack's consumer stores. Defaults to the service name.
func ResolveSource(serviceName string) string {
	if s := strings.TrimSpace(serviceName); s != "" {
		return s
	}
	return DefaultServiceName
}

// ResolveGroupID picks the consumer group. An explicit id wins; otherwise it
// is "<source>-activity-log", so two stacks never share a group (a shared
// group would split partitions and each stack would miss events).
func ResolveGroupID(explicit, source string) string {
	if g := strings.TrimSpace(explicit); g != "" {
		return g
	}
	return source + "-activity-log"
}

// EnvTopicPrefix returns KAFKA_TOPIC_PREFIX.
func EnvTopicPrefix() string { return os.Getenv(TopicPrefixEnv) }

// EnvServiceName returns KAFKA_SERVICE_NAME.
func EnvServiceName() string { return os.Getenv(ServiceNameEnv) }
