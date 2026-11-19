// Package datasource holds the SSL/TLS rules for external PostgreSQL data sources.
//
// The modes follow libpq semantics:
//   - disable     no TLS
//   - require     TLS is mandatory, the server certificate is NOT verified
//   - verify-ca   TLS is mandatory, the certificate chain must verify against the CA pool
//   - verify-full verify-ca plus the certificate must match the host name
//
// For verify-ca / verify-full an empty CA certificate means the system root pool is used.
// A server that does not accept SSLRequest is always an error: there is no plaintext fallback.
package datasource

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	SSLModeDisable    = "disable"
	SSLModeRequire    = "require"
	SSLModeVerifyCA   = "verify-ca"
	SSLModeVerifyFull = "verify-full"

	// MaxRootCertBytes bounds the CA bundle accepted from the API.
	MaxRootCertBytes = 64 * 1024

	// sslRequestCode is the PostgreSQL SSLRequest magic number (1234.5679).
	sslRequestCode = 80877103
)

// NormalizeSSLMode returns the effective sslmode. When mode is empty the legacy
// boolean flag decides: true -> require, false -> disable. A non-empty mode is
// authoritative and the boolean is ignored.
func NormalizeSSLMode(sslEnabled bool, mode string) (string, error) {
	m := strings.ToLower(strings.TrimSpace(mode))
	if m == "" {
		if sslEnabled {
			return SSLModeRequire, nil
		}
		return SSLModeDisable, nil
	}
	if !isKnownMode(m) {
		return "", fmt.Errorf("unknown ssl_mode %q: allowed values are disable, require, verify-ca, verify-full", mode)
	}
	return m, nil
}

// ValidateSSLConfig checks the mode and optional CA certificate PEM together.
func ValidateSSLConfig(mode, rootCertPEM string) error {
	if !isKnownMode(mode) {
		return fmt.Errorf("unknown ssl_mode %q", mode)
	}
	if rootCertPEM == "" {
		return nil
	}
	if mode == SSLModeDisable || mode == SSLModeRequire {
		return fmt.Errorf("ssl_root_cert only applies to verify-ca and verify-full, not %s", mode)
	}
	if len(rootCertPEM) > MaxRootCertBytes {
		return fmt.Errorf("ssl_root_cert exceeds %d bytes", MaxRootCertBytes)
	}
	if _, err := parseRootPool(rootCertPEM); err != nil {
		return err
	}
	return nil
}

// BuildTLSConfig returns the client TLS config for a mode, or nil for disable.
// host is used as ServerName for verify-full.
func BuildTLSConfig(host, mode, rootCertPEM string) (*tls.Config, error) {
	if err := ValidateSSLConfig(mode, rootCertPEM); err != nil {
		return nil, err
	}
	switch mode {
	case SSLModeDisable:
		return nil, nil
	case SSLModeRequire:
		return &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}, nil //nolint:gosec // libpq "require" semantics
	}

	var pool *x509.CertPool // nil = system roots
	if rootCertPEM != "" {
		p, err := parseRootPool(rootCertPEM)
		if err != nil {
			return nil, err
		}
		pool = p
	}

	if mode == SSLModeVerifyFull {
		return &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host, RootCAs: pool}, nil
	}

	// verify-ca: check the chain ourselves, skip only the host name check.
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, //nolint:gosec // chain is verified in VerifyConnection
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return errors.New("server presented no certificate")
			}
			inter := x509.NewCertPool()
			for _, c := range cs.PeerCertificates[1:] {
				inter.AddCert(c)
			}
			_, err := cs.PeerCertificates[0].Verify(x509.VerifyOptions{Roots: pool, Intermediates: inter})
			return err
		},
	}, nil
}

// BuildPostgresDSN returns a libpq keyword/value connection string that carries
// the sslmode. The CA certificate is not part of the DSN (it is PEM, not a file);
// it is applied through the TLS config from BuildTLSConfig.
func BuildPostgresDSN(host string, port int, database, user, password, mode string) (string, error) {
	if !isKnownMode(mode) {
		return "", fmt.Errorf("unknown ssl_mode %q", mode)
	}
	if host == "" {
		return "", errors.New("host is required")
	}
	if port <= 0 {
		port = 5432
	}
	parts := []string{
		"host=" + quoteDSNValue(host),
		"port=" + strconv.Itoa(port),
		"dbname=" + quoteDSNValue(database),
		"user=" + quoteDSNValue(user),
		"password=" + quoteDSNValue(password),
		"sslmode=" + mode,
	}
	return strings.Join(parts, " "), nil
}

// ResolveDockerFallbackAddrs returns fallback host:port candidates when master-service runs inside
// a Docker container (/.dockerenv exists) and the user configured a loopback or Docker-internal host.
func ResolveDockerFallbackAddrs(host string, port int) []string {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return nil
	}
	if port <= 0 {
		port = 5432
	}
	h := strings.ToLower(strings.TrimSpace(host))
	var out []string
	switch h {
	case "localhost", "127.0.0.1", "::1":
		out = append(out, net.JoinHostPort("host.docker.internal", strconv.Itoa(port)))
		if port == 5432 {
			out = append(out, "host.docker.internal:5434", "postgres:5432")
		}
	case "host.docker.internal":
		if port == 5432 {
			out = append(out, "host.docker.internal:5434", "postgres:5432")
		}
	case "postgres-cbs-sim", "auditsphere_cbs":
		out = append(out, "host.docker.internal:5434")
	}
	return out
}

func dialWithDockerFallback(ctx context.Context, host string, port int) (net.Conn, string, error) {
	if port <= 0 {
		port = 5432
	}
	primaryAddr := net.JoinHostPort(host, strconv.Itoa(port))
	var d net.Dialer
	raw, err := d.DialContext(ctx, "tcp", primaryAddr)
	if err == nil {
		return raw, primaryAddr, nil
	}
	for _, cand := range ResolveDockerFallbackAddrs(host, port) {
		if c, fErr := d.DialContext(ctx, "tcp", cand); fErr == nil {
			return c, cand, nil
		}
	}
	return nil, primaryAddr, fmt.Errorf("connect to %s: %w", primaryAddr, err)
}

// DialPostgresTCP checks plain TCP reachability (with Docker loopback fallback when inside a container).
func DialPostgresTCP(ctx context.Context, host string, port int, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	c, _, err := dialWithDockerFallback(ctx, host, port)
	if err != nil {
		return err
	}
	return c.Close()
}

// ProbePostgresTLS performs the PostgreSQL SSLRequest exchange and a TLS handshake
// using the config from BuildTLSConfig. It does not authenticate. Mode disable is rejected.
func ProbePostgresTLS(ctx context.Context, host string, port int, mode, rootCertPEM string, timeout time.Duration) error {
	if mode == SSLModeDisable {
		return errors.New("ssl probe requires an ssl_mode other than disable")
	}
	tlsCfg, err := BuildTLSConfig(host, mode, rootCertPEM)
	if err != nil {
		return err
	}
	if port <= 0 {
		port = 5432
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	primaryAddr := net.JoinHostPort(host, strconv.Itoa(port))
	primaryErr := probePostgresTLSAddr(ctx, primaryAddr, tlsCfg, mode, timeout)
	if primaryErr == nil {
		return nil
	}

	for _, candAddr := range ResolveDockerFallbackAddrs(host, port) {
		if err := probePostgresTLSAddr(ctx, candAddr, tlsCfg, mode, timeout); err == nil {
			return nil
		}
	}
	return primaryErr
}

func probePostgresTLSAddr(parentCtx context.Context, addr string, tlsCfg *tls.Config, mode string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	var d net.Dialer
	raw, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer raw.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = raw.SetDeadline(deadline)
	}

	req := make([]byte, 8)
	binary.BigEndian.PutUint32(req[0:4], 8)
	binary.BigEndian.PutUint32(req[4:8], sslRequestCode)
	if _, err := raw.Write(req); err != nil {
		return fmt.Errorf("send SSLRequest to %s: %w", addr, err)
	}
	resp := make([]byte, 1)
	if _, err := io.ReadFull(raw, resp); err != nil {
		return fmt.Errorf("read SSLRequest response from %s: %w", addr, err)
	}
	switch resp[0] {
	case 'S':
	case 'N':
		return fmt.Errorf("server %s does not accept SSL (ssl_mode=%s); refusing to continue without encryption", addr, mode)
	default:
		return fmt.Errorf("unexpected SSLRequest response %q from %s", resp[0], addr)
	}

	tc := tls.Client(raw, tlsCfg.Clone())
	if err := tc.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("TLS handshake with %s (ssl_mode=%s) failed: %w", addr, mode, err)
	}
	return tc.Close()
}

func isKnownMode(m string) bool {
	switch m {
	case SSLModeDisable, SSLModeRequire, SSLModeVerifyCA, SSLModeVerifyFull:
		return true
	}
	return false
}

func parseRootPool(pemData string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(pemData)) {
		return nil, errors.New("ssl_root_cert contains no valid PEM certificate")
	}
	return pool, nil
}

func quoteDSNValue(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + r.Replace(s) + "'"
}
