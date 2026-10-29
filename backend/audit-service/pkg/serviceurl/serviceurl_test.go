package serviceurl

import "testing"

func TestResolve(t *testing.T) {
	cases := []struct {
		name, explicit, service, want string
	}{
		{"explicit wins over tenant slug", " http://example:9000/ ", "audit-service-bai", "http://example:9000"},
		{"explicit, several trailing slashes", "http://localhost:8003//", "", "http://localhost:8003"},
		{"whitespace-only explicit is unset", "   ", "", "http://master-service:8002"},
		{"tenant slug", "", "audit-service-bai", "http://master-service-bai:8002"},
		{"tenant slug, padded and upper case", "", "  Audit-Service-BAI ", "http://master-service-bai:8002"},
		{"tenant slug with digits and hyphen", "", "audit-service-acme-2", "http://master-service-acme-2:8002"},
		{"prod: unset", "", "", "http://master-service:8002"},
		{"prod: config default name", "", "audit-service", "http://master-service:8002"},
		{"another service's name is not a tenant", "", "risk-service-bai", "http://master-service:8002"},
		{"prefix only, empty slug", "", "audit-service-", ""},
		{"slug with a port", "", "audit-service-bai:9999", ""},
		{"slug with a path", "", "audit-service-bai/x", ""},
		{"slug with userinfo", "", "audit-service-evil@host", ""},
		{"slug with a dot", "", "audit-service-bai.example.com", ""},
		{"slug ending in hyphen", "", "audit-service-bai-", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Resolve(tc.explicit, tc.service, "master-service", MasterServicePort); got != tc.want {
				t.Fatalf("Resolve(%q, %q) = %q, want %q", tc.explicit, tc.service, got, tc.want)
			}
		})
	}
}

func TestMasterAndAuthFromEnv(t *testing.T) {
	t.Run("tenant", func(t *testing.T) {
		t.Setenv(MasterServiceURLEnv, "")
		t.Setenv(AuthServiceURLEnv, "")
		t.Setenv(ServiceNameEnv, "audit-service-bai")
		if got := Master(); got != "http://master-service-bai:8002" {
			t.Fatalf("Master() = %q", got)
		}
		if got := Auth(); got != "http://auth-service-bai:8001" {
			t.Fatalf("Auth() = %q", got)
		}
	})
	t.Run("prod default", func(t *testing.T) {
		t.Setenv(MasterServiceURLEnv, "")
		t.Setenv(AuthServiceURLEnv, "")
		t.Setenv(ServiceNameEnv, "")
		if got := Master(); got != "http://master-service:8002" {
			t.Fatalf("Master() = %q", got)
		}
		if got := Auth(); got != "http://auth-service:8001" {
			t.Fatalf("Auth() = %q", got)
		}
	})
	t.Run("risk: tenant, prod and explicit", func(t *testing.T) {
		t.Setenv(RiskServiceURLEnv, "")
		t.Setenv(ServiceNameEnv, "audit-service-bai")
		if got := Risk(); got != "http://risk-service-bai:8002" {
			t.Fatalf("tenant Risk() = %q", got)
		}
		t.Setenv(ServiceNameEnv, "")
		if got := Risk(); got != "http://risk-service:8002" {
			t.Fatalf("prod Risk() = %q", got)
		}
		t.Setenv(RiskServiceURLEnv, "http://localhost:8004/")
		t.Setenv(ServiceNameEnv, "audit-service-bai")
		if got := Risk(); got != "http://localhost:8004" {
			t.Fatalf("explicit Risk() = %q", got)
		}
		t.Setenv(RiskServiceURLEnv, "")
		t.Setenv(ServiceNameEnv, "audit-service-bai.example.com")
		if got := Risk(); got != "" {
			t.Fatalf("malformed tenant slug Risk() = %q, want empty", got)
		}
	})
	t.Run("explicit env per target", func(t *testing.T) {
		t.Setenv(MasterServiceURLEnv, "http://localhost:8003/")
		t.Setenv(AuthServiceURLEnv, "http://localhost:8001")
		t.Setenv(ServiceNameEnv, "audit-service-bai")
		if got := Master(); got != "http://localhost:8003" {
			t.Fatalf("Master() = %q", got)
		}
		if got := Auth(); got != "http://localhost:8001" {
			t.Fatalf("Auth() = %q", got)
		}
	})
}
