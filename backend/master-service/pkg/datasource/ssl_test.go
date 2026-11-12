package datasource

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNormalizeSSLMode(t *testing.T) {
	cases := []struct {
		name    string
		enabled bool
		mode    string
		want    string
		wantErr bool
	}{
		{"legacy true", true, "", SSLModeRequire, false},
		{"legacy false", false, "", SSLModeDisable, false},
		{"explicit verify-full wins over flag", false, "verify-full", SSLModeVerifyFull, false},
		{"case and spaces", true, "  Verify-CA ", SSLModeVerifyCA, false},
		{"unknown value rejected", true, "prefer", "", true},
		{"allow is not libpq-safe here", true, "allow", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeSSLMode(tc.enabled, tc.mode)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("mode = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateSSLConfig(t *testing.T) {
	ca := newTestCA(t)
	caPEM := ca.pem
	cases := []struct {
		name    string
		mode    string
		cert    string
		wantErr bool
	}{
		{"verify-full with system pool", SSLModeVerifyFull, "", false},
		{"verify-ca with system pool", SSLModeVerifyCA, "", false},
		{"verify-full with CA", SSLModeVerifyFull, caPEM, false},
		{"require with cert rejected", SSLModeRequire, caPEM, true},
		{"disable with cert rejected", SSLModeDisable, caPEM, true},
		{"garbage PEM rejected", SSLModeVerifyCA, "not a certificate", true},
		{"unknown mode rejected", "verify_full", "", true},
		{"oversize cert rejected", SSLModeVerifyCA, strings.Repeat("A", MaxRootCertBytes+1), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSSLConfig(tc.mode, tc.cert)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestBuildTLSConfig(t *testing.T) {
	caPEM := newTestCA(t).pem

	if cfg, err := BuildTLSConfig("db", SSLModeDisable, ""); err != nil || cfg != nil {
		t.Fatalf("disable: cfg=%v err=%v, want nil,nil", cfg, err)
	}

	cfg, err := BuildTLSConfig("db.example", SSLModeRequire, "")
	if err != nil || cfg == nil || !cfg.InsecureSkipVerify {
		t.Fatalf("require: want TLS config without verification, got %+v err=%v", cfg, err)
	}

	cfg, err = BuildTLSConfig("db.example", SSLModeVerifyFull, caPEM)
	if err != nil || cfg == nil {
		t.Fatalf("verify-full: err=%v", err)
	}
	if cfg.InsecureSkipVerify || cfg.ServerName != "db.example" || cfg.RootCAs == nil {
		t.Fatalf("verify-full: want ServerName and RootCAs set, no skip-verify; got skip=%v sn=%q pool=%v",
			cfg.InsecureSkipVerify, cfg.ServerName, cfg.RootCAs != nil)
	}

	cfg, err = BuildTLSConfig("db.example", SSLModeVerifyFull, "")
	if err != nil || cfg.RootCAs != nil {
		t.Fatalf("verify-full without CA: want nil RootCAs (system pool), got err=%v pool=%v", err, cfg.RootCAs != nil)
	}

	cfg, err = BuildTLSConfig("db.example", SSLModeVerifyCA, caPEM)
	if err != nil || cfg == nil || cfg.VerifyConnection == nil || cfg.ServerName != "" {
		t.Fatalf("verify-ca: want VerifyConnection and no ServerName, got err=%v cfg=%+v", err, cfg)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("verify-ca: expected InsecureSkipVerify with custom VerifyConnection")
	}
}

func TestBuildPostgresDSN(t *testing.T) {
	dsn, err := BuildPostgresDSN("db.example", 0, "gl", "reader", `p'ss\word`, SSLModeVerifyFull)
	if err != nil {
		t.Fatal(err)
	}
	want := `host='db.example' port=5432 dbname='gl' user='reader' password='p\'ss\\word' sslmode=verify-full`
	if dsn != want {
		t.Fatalf("dsn = %s\nwant  %s", dsn, want)
	}

	if _, err := BuildPostgresDSN("db", 5432, "gl", "u", "p", "allow"); err == nil {
		t.Fatal("unknown mode must fail")
	}
	if _, err := BuildPostgresDSN("", 5432, "gl", "u", "p", SSLModeRequire); err == nil {
		t.Fatal("empty host must fail")
	}
}

func TestProbePostgresTLS(t *testing.T) {
	ca := newTestCA(t)
	caPEM := ca.pem
	// Server cert SAN is a DNS name only, so verify-full (dialing 127.0.0.1) must fail while verify-ca passes.
	serverCfg := ca.serverTLS(t, []string{"db.internal"}, nil)
	otherCAPEM := newTestCA(t).pem

	cases := []struct {
		name    string
		mode    string
		rootPEM string
		server  string // "tls" = accept SSLRequest and speak TLS, "plain" = answer N
		wantErr string // substring; empty = success
	}{
		{"require accepts untrusted cert", SSLModeRequire, "", "tls", ""},
		{"verify-ca with trusted CA", SSLModeVerifyCA, caPEM, "tls", ""},
		{"verify-ca with wrong CA", SSLModeVerifyCA, otherCAPEM, "tls", "certificate"},
		{"verify-ca with system pool", SSLModeVerifyCA, "", "tls", "certificate"},
		{"verify-full host mismatch", SSLModeVerifyFull, caPEM, "tls", "certificate"},
		{"server refuses SSL", SSLModeRequire, "", "plain", "does not accept SSL"},
		{"verify-ca refuses plaintext too", SSLModeVerifyCA, caPEM, "plain", "does not accept SSL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			addr, stop := startFakePostgres(t, tc.server, serverCfg)
			defer stop()
			host, portStr, _ := net.SplitHostPort(addr)
			port, _ := strconv.Atoi(portStr)

			err := ProbePostgresTLS(context.Background(), host, port, tc.mode, tc.rootPEM, 3*time.Second)
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
}

func TestProbeRejectsDisable(t *testing.T) {
	if err := ProbePostgresTLS(context.Background(), "127.0.0.1", 5432, SSLModeDisable, "", time.Second); err == nil {
		t.Fatal("disable must not be probed as TLS")
	}
}

// startFakePostgres answers the SSLRequest like a PostgreSQL server.
// mode "tls" replies 'S' and performs a server-side TLS handshake; "plain" replies 'N'.
func startFakePostgres(t *testing.T, mode string, serverCfg *tls.Config) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(3 * time.Second))
				buf := make([]byte, 8)
				if _, err := io.ReadFull(c, buf); err != nil {
					return
				}
				if binary.BigEndian.Uint32(buf[4:8]) != sslRequestCode {
					return
				}
				if mode == "plain" {
					_, _ = c.Write([]byte{'N'})
					return
				}
				_, _ = c.Write([]byte{'S'})
				tc := tls.Server(c, serverCfg)
				_ = tc.Handshake() // client-side errors surface on its side
			}(c)
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

type testCA struct {
	pem  string
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

func newTestCA(t *testing.T) testCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return testCA{pem: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), cert: cert, key: key}
}

// serverTLS issues a server certificate signed by ca with the given SANs.
func (ca testCA) serverTLS(t *testing.T, dns []string, ips []net.IP) *tls.Config {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano() + 1),
		Subject:      pkix.Name{CommonName: "db"},
		DNSNames:     dns,
		IPAddresses:  ips,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
	}
}
