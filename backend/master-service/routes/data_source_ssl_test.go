package routes

import (
	"strings"
	"testing"

	"master-service/models"
)

const testRootPEM = `-----BEGIN CERTIFICATE-----
MIIBeTCCAR+gAwIBAgIUM8qWJbWKFfQ3n2E8EnLxXilUsCwwCgYIKoZIzj0EAwIw
EjEQMA4GA1UEAwwHdGVzdC1jYTAeFw0yNjExMDgxMTE5MTVaFw0zNjExMDUxMTE5
MTVaMBIxEDAOBgNVBAMMB3Rlc3QtY2EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AATLa68KKp6N0AVAMdH0kqoAyCx4FUwmlJN9IoJrhndDQrlJDZBZmswARUrZspZ0
hNZCtNGLW/y5wLQhliG7KgfZo1MwUTAdBgNVHQ4EFgQU1RInMTkosMtyph2WAY3Z
59+GMXEwHwYDVR0jBBgwFoAU1RInMTkosMtyph2WAY3Z59+GMXEwDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiEA39ydqEULsllt7IpZ2dck7FIo8qar
nVTL3cMPV5pPDMUCIC7xonh2nDnUix9rSb5+iN2Fn0cQXMw0NxF0LJrQF5Bg
-----END CERTIFICATE-----`

func TestApplyDataSourceSSL_Create(t *testing.T) {
	cases := []struct {
		name     string
		conn     models.DataSourceConnection
		wantMode string
		wantSSL  bool
		wantErr  bool
	}{
		{"legacy ssl true", models.DataSourceConnection{SSL: true}, "require", true, false},
		{"legacy ssl false", models.DataSourceConnection{SSL: false}, "disable", false, false},
		{"explicit verify-full keeps ssl on", models.DataSourceConnection{SSL: false, SSLMode: "verify-full"}, "verify-full", true, false},
		{"unknown mode", models.DataSourceConnection{SSL: true, SSLMode: "prefer"}, "", false, true},
		{"cert with require rejected", models.DataSourceConnection{SSL: true, SSLMode: "require", SSLRootCert: testRootPEM}, "", false, true},
		{"bad PEM rejected", models.DataSourceConnection{SSLMode: "verify-ca", SSLRootCert: "junk"}, "", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn := tc.conn
			err := applyDataSourceSSL(&conn, conn.SSLRootCert)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if conn.SSLMode != tc.wantMode || conn.SSL != tc.wantSSL {
				t.Fatalf("got mode=%q ssl=%v, want mode=%q ssl=%v", conn.SSLMode, conn.SSL, tc.wantMode, tc.wantSSL)
			}
		})
	}
}

func TestApplyDataSourceSSL_Update(t *testing.T) {
	stored := models.DataSourceConnection{SSL: true, SSLMode: "verify-full", SSLRootCert: testRootPEM}

	// Save from the settings UI, which does not resend the CA: keep the stored cert.
	existing := stored
	existing.SSLMode = "verify-full"
	existing.SSL = true
	if err := applyDataSourceSSL(&existing, ""); err != nil {
		t.Fatalf("keep cert: %v", err)
	}
	if existing.SSLRootCert != testRootPEM {
		t.Fatal("stored CA certificate was dropped on an update that did not send one")
	}

	// Switching to require drops the stale CA.
	existing = stored
	existing.SSLMode = "require"
	existing.SSL = true
	if err := applyDataSourceSSL(&existing, ""); err != nil {
		t.Fatalf("switch to require: %v", err)
	}
	if existing.SSLRootCert != "" || existing.SSLMode != "require" {
		t.Fatalf("stale cert kept for require: mode=%q cert=%q", existing.SSLMode, existing.SSLRootCert)
	}

	// Explicit new cert with require is an error, not silently ignored.
	existing = stored
	existing.SSLMode = "require"
	existing.SSL = true
	if err := applyDataSourceSSL(&existing, testRootPEM); err == nil || !strings.Contains(err.Error(), "verify-ca") {
		t.Fatalf("explicit cert with require: err = %v", err)
	}
}
