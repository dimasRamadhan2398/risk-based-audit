package config

import "testing"

func TestMasterServiceBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		serviceName string
		want        string
	}{
		{"shared stack default", "", "", "http://master-service:8002"},
		{"shared stack with plain service name", "", "risk-service", "http://master-service:8002"},
		{"tenant stack derives its own master-service", "", "risk-service-bai", "http://master-service-bai:8002"},
		{"explicit url wins over tenant derivation", "http://localhost:8003/", "risk-service-bai", "http://localhost:8003"},
		{"unrelated service name is not treated as a tenant", "", "master-service-bai", "http://master-service:8002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KAFKA_SERVICE_NAME", tt.serviceName)
			got := MasterServiceConfig{URL: tt.url}.BaseURL()
			if got != tt.want {
				t.Errorf("BaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
