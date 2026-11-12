package datasource

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// Integration test against real PostgreSQL containers (one with ssl=on and a throwaway CA, one
// without TLS). It is skipped with -short or when docker / the postgres image is unavailable.
// Override the image with DATASOURCE_TEST_PG_IMAGE (default postgres:16-alpine).

func dockerOrSkip(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon not reachable")
	}
	img := os.Getenv("DATASOURCE_TEST_PG_IMAGE")
	if img == "" {
		img = "postgres:16-alpine"
	}
	if err := exec.Command("docker", "image", "inspect", img).Run(); err != nil {
		t.Skipf("image %s not present locally (not pulling in tests)", img)
	}
	return img
}

// startPG runs a throwaway container and returns its published 127.0.0.1 port. The container is
// removed with t.Cleanup. certDir, when set, enables ssl using server.crt / server.key from it.
func startPG(t *testing.T, img, certDir string) int {
	t.Helper()
	name := "ds-ssl-test-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	args := []string{"run", "-d", "--rm", "--name", name, "-e", "POSTGRES_PASSWORD=pw", "-p", "127.0.0.1::5432"}
	if certDir == "" {
		args = append(args, img)
	} else {
		// Postgres requires a 0600 key owned by the postgres user, so copy the mounted files first.
		script := `cp /certs/server.crt /certs/server.key /var/lib/postgresql/ && ` +
			`chown postgres:postgres /var/lib/postgresql/server.* && chmod 600 /var/lib/postgresql/server.key && ` +
			`exec docker-entrypoint.sh postgres -c ssl=on -c ssl_cert_file=/var/lib/postgresql/server.crt -c ssl_key_file=/var/lib/postgresql/server.key`
		args = append(args, "-v", certDir+":/certs:ro", "--entrypoint", "sh", img, "-c", script)
	}
	if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		t.Fatalf("docker run: %v: %s", err, out)
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })

	var port int
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		// initdb starts a temporary server first; only the final one accepts TCP on 127.0.0.1.
		if exec.Command("docker", "exec", name, "pg_isready", "-h", "127.0.0.1", "-U", "postgres").Run() == nil {
			out, err := exec.Command("docker", "port", name, "5432/tcp").Output()
			if err != nil {
				t.Fatalf("docker port: %v", err)
			}
			line := strings.Fields(string(out))[0]
			_, p, _ := net.SplitHostPort(line)
			port, _ = strconv.Atoi(p)
			return port
		}
		time.Sleep(500 * time.Millisecond)
	}
	logs, _ := exec.Command("docker", "logs", name).CombinedOutput()
	t.Fatalf("postgres did not become ready: %s", logs)
	return 0
}

func writeServerCert(t *testing.T, ca testCA, dir string) {
	t.Helper()
	cfg := ca.serverTLS(t, []string{"localhost"}, []net.IP{net.ParseIP("127.0.0.1")})
	crt := cfg.Certificates[0]
	keyDER, err := x509.MarshalPKCS8PrivateKey(crt.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Certificate[0]})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(filepath.Join(dir, "server.crt"), certPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "server.key"), keyPEM, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestProbePostgresTLS_RealPostgres(t *testing.T) {
	img := dockerOrSkip(t)
	ca := newTestCA(t)
	otherCA := newTestCA(t)
	dir := t.TempDir()
	writeServerCert(t, ca, dir)

	sslPort := startPG(t, img, dir)
	plainPort := startPG(t, img, "")
	ctx := context.Background()

	cases := []struct {
		name    string
		host    string
		port    int
		mode    string
		rootPEM string
		wantErr string
	}{
		{"require", "127.0.0.1", sslPort, SSLModeRequire, "", ""},
		{"verify-ca trusted CA", "127.0.0.1", sslPort, SSLModeVerifyCA, ca.pem, ""},
		{"verify-full IP SAN", "127.0.0.1", sslPort, SSLModeVerifyFull, ca.pem, ""},
		{"verify-full DNS SAN", "localhost", sslPort, SSLModeVerifyFull, ca.pem, ""},
		{"verify-ca wrong CA", "127.0.0.1", sslPort, SSLModeVerifyCA, otherCA.pem, "certificate"},
		{"verify-ca system pool", "127.0.0.1", sslPort, SSLModeVerifyCA, "", "certificate"},
		{"require on server without TLS", "127.0.0.1", plainPort, SSLModeRequire, "", "does not accept SSL"},
		{"verify-full on server without TLS", "127.0.0.1", plainPort, SSLModeVerifyFull, ca.pem, "does not accept SSL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ProbePostgresTLS(ctx, tc.host, tc.port, tc.mode, tc.rootPEM, 5*time.Second)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want substring %q", err, tc.wantErr)
			}
		})
	}

	// A real authenticated session using the DSN + TLS config this package builds.
	t.Run("pgx session over verify-full", func(t *testing.T) {
		dsn, err := BuildPostgresDSN("localhost", sslPort, "postgres", "postgres", "pw", SSLModeVerifyFull)
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := pgx.ParseConfig(dsn)
		if err != nil {
			t.Fatal(err)
		}
		tlsCfg, err := BuildTLSConfig("localhost", SSLModeVerifyFull, ca.pem)
		if err != nil {
			t.Fatal(err)
		}
		cfg.TLSConfig = tlsCfg
		cfg.Fallbacks = nil
		c, err := pgx.ConnectConfig(ctx, cfg)
		if err != nil {
			t.Fatalf("connect: %v", err)
		}
		defer c.Close(ctx)
		var ssl bool
		if err := c.QueryRow(ctx, "SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid()").Scan(&ssl); err != nil || !ssl {
			t.Fatalf("session not encrypted: ssl=%v err=%v", ssl, err)
		}
	})
}
