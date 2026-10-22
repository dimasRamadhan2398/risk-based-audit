package config

import "testing"

func TestMasterServiceBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		service string
		want    string
	}{
		{"explicit url wins", "http://m.internal:9000/", "auth-service-bai", "http://m.internal:9000"},
		{"tenant derives its own master", "", "auth-service-bai", "http://master-service-bai:8002"},
		{"shared stack uses bare alias", "", "auth-service", "http://master-service:8002"},
		{"unset name uses bare alias", "", "", "http://master-service:8002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KAFKA_SERVICE_NAME", tt.service)
			got := MasterServiceConfig{URL: tt.url}.BaseURL()
			if got != tt.want {
				t.Fatalf("BaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
