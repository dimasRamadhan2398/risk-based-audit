// Package serviceurl picks the base URL audit-service uses to reach a sibling
// service (master-service, auth-service) of its own stack.
//
// Tenant stacks share the rb_audit_network bridge with the control plane, so
// the bare alias "master-service" resolves to the *control plane's*
// master-service from inside a tenant container. Tenant services are named
// "<service>-<slug>" (scripts/templates/docker-compose.tenant.yml.tpl) and set
// KAFKA_SERVICE_NAME to that name, so the sibling is "<target>-<slug>". This
// mirrors risk-service's derivedMasterServiceURL and auth-service's
// MasterServiceConfig.BaseURL.
package serviceurl

import (
	"os"
	"regexp"
	"strings"
)

const (
	// selfPrefix is this service's name prefix in a tenant stack
	selfPrefix = "audit-service-"

	// Container ports inside the docker network (control plane and tenants)
	MasterServicePort = "8002"
	AuthServicePort   = "8001"

	// Env vars
	MasterServiceURLEnv = "MASTER_SERVICE_URL"
	AuthServiceURLEnv   = "AUTH_SERVICE_URL"
	ServiceNameEnv      = "KAFKA_SERVICE_NAME"
)

// slugPattern is what a tenant slug may look like in a docker alias. The
// generator is stricter ([a-z][a-z0-9]{1,30}); this only rules out values that
// could change the host part of the URL (":", "/", "@", ".", spaces).
var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Resolve returns the base URL (no trailing slash) of target listening on port.
//
//   - explicit (the *_SERVICE_URL env value) wins, trimmed, trailing "/" removed
//   - serviceName "audit-service-<slug>" gives http://<target>-<slug>:<port>
//   - anything else (unset, "audit-service", another service's name) is the
//     control plane / prod: http://<target>:<port>
//
// It returns "" when serviceName claims to be a tenant ("audit-service-...")
// but the slug is malformed: falling back to the shared service there would
// send a tenant's calls to the control plane, so the caller must treat ""
// as "not configured" and not call out at all.
func Resolve(explicit, serviceName, target, port string) string {
	if u := strings.TrimRight(strings.TrimSpace(explicit), "/"); u != "" {
		return u
	}
	name := strings.ToLower(strings.TrimSpace(serviceName))
	if strings.HasPrefix(name, selfPrefix) {
		slug := strings.TrimPrefix(name, selfPrefix)
		if !slugPattern.MatchString(slug) {
			return ""
		}
		return "http://" + target + "-" + slug + ":" + port
	}
	return "http://" + target + ":" + port
}

// Master returns this stack's master-service base URL from the environment
func Master() string {
	return Resolve(os.Getenv(MasterServiceURLEnv), os.Getenv(ServiceNameEnv), "master-service", MasterServicePort)
}

// Auth returns this stack's auth-service base URL from the environment
func Auth() string {
	return Resolve(os.Getenv(AuthServiceURLEnv), os.Getenv(ServiceNameEnv), "auth-service", AuthServicePort)
}
