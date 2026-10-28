// Package avscan checks uploaded files for malware with a ClamAV daemon
// (clamd) reached over TCP, using the INSTREAM command.
//
// Configuration comes from the environment:
//
//	CLAMAV_ADDRESS  host:port of clamd (default "clamav:3310", the compose sidecar)
//	CLAMAV_MODE     required (default) | optional | off
//	CLAMAV_TIMEOUT  per-scan timeout as a Go duration (default "30s")
//
// Mode "required" is fail-closed: when clamd cannot be reached, or does not
// give a clear answer, the upload is refused. "optional" lets the upload
// through with a logged warning when clamd is unreachable (still refusing
// infected files). "off" disables scanning entirely and is meant only for
// local development and tests.
package avscan

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Mode decides what happens when the scanner cannot give an answer.
type Mode string

const (
	// ModeRequired refuses the upload when clamd is unreachable (fail-closed).
	ModeRequired Mode = "required"
	// ModeOptional accepts the upload when clamd is unreachable (fail-open).
	ModeOptional Mode = "optional"
	// ModeOff skips scanning.
	ModeOff Mode = "off"
)

const (
	DefaultAddress = "clamav:3310"
	DefaultTimeout = 30 * time.Second
	chunkSize      = 64 << 10
)

// ErrUnavailable means clamd could not be reached or gave no usable answer.
var ErrUnavailable = errors.New("malware scanner unavailable")

// InfectedError is returned for a file that clamd reported as infected.
type InfectedError struct {
	Signature string
}

func (e *InfectedError) Error() string {
	return "malware detected: " + e.Signature
}

// IsInfected reports whether err is (or wraps) an InfectedError.
func IsInfected(err error) bool {
	var ie *InfectedError
	return errors.As(err, &ie)
}

// Scanner scans a stream. A clean stream returns nil; an infected one returns
// an *InfectedError; any other error means the scan did not complete.
type Scanner interface {
	Scan(ctx context.Context, r io.Reader) error
}

// Clamd talks to clamd over TCP.
type Clamd struct {
	Address string
	Timeout time.Duration
}

// Scan streams r to clamd with INSTREAM and interprets the reply.
func (c *Clamd) Scan(ctx context.Context, r io.Reader) error {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.Address)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}

	w := bufio.NewWriterSize(conn, chunkSize+4)
	if _, err := w.WriteString("zINSTREAM\x00"); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	buf := make([]byte, chunkSize)
	var size [4]byte
	for {
		n, rerr := r.Read(buf)
		if n > 0 {
			binary.BigEndian.PutUint32(size[:], uint32(n))
			if _, err := w.Write(size[:]); err != nil {
				return fmt.Errorf("%w: %v", ErrUnavailable, err)
			}
			if _, err := w.Write(buf[:n]); err != nil {
				return fmt.Errorf("%w: %v", ErrUnavailable, err)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fmt.Errorf("read upload for scanning: %w", rerr)
		}
	}
	binary.BigEndian.PutUint32(size[:], 0)
	if _, err := w.Write(size[:]); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	reply, err := bufio.NewReader(conn).ReadString(0)
	if err != nil && !(errors.Is(err, io.EOF) && reply != "") {
		return fmt.Errorf("%w: reading reply: %v", ErrUnavailable, err)
	}
	return parseReply(reply)
}

// parseReply interprets "stream: OK", "stream: <sig> FOUND" and "... ERROR".
func parseReply(reply string) error {
	reply = strings.TrimSpace(strings.TrimRight(reply, "\x00"))
	switch {
	case strings.HasSuffix(reply, " OK"):
		return nil
	case strings.HasSuffix(reply, " FOUND"):
		sig := strings.TrimSuffix(reply, " FOUND")
		if i := strings.Index(sig, ": "); i >= 0 {
			sig = sig[i+2:]
		}
		return &InfectedError{Signature: strings.TrimSpace(sig)}
	default:
		// "INSTREAM size limit exceeded. ERROR" and anything unexpected: the
		// file was not (fully) scanned.
		return fmt.Errorf("%w: clamd replied %q", ErrUnavailable, reply)
	}
}

// Ping checks that clamd answers PING with PONG.
func (c *Clamd) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.Address)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}
	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	reply, _ := bufio.NewReader(conn).ReadString(0)
	if strings.TrimRight(reply, "\x00\n") != "PONG" {
		return fmt.Errorf("%w: clamd replied %q to PING", ErrUnavailable, reply)
	}
	return nil
}

// Guard applies the mode to a scanner.
type Guard struct {
	Scanner Scanner
	Mode    Mode
	Logf    func(format string, args ...any)
}

func (g *Guard) logf(format string, args ...any) {
	if g.Logf != nil {
		g.Logf(format, args...)
		return
	}
	log.Printf(format, args...)
}

// CheckBytes scans data. label identifies the upload in logs (never the file
// content). It returns nil when the file may be stored, an *InfectedError for
// malware, or an error wrapping ErrUnavailable when the scan failed and the
// mode is fail-closed.
func (g *Guard) CheckBytes(ctx context.Context, label string, data []byte) error {
	return g.Check(ctx, label, bytes.NewReader(data))
}

// Check is CheckBytes for a stream.
func (g *Guard) Check(ctx context.Context, label string, r io.Reader) error {
	if g == nil || g.Mode == ModeOff || g.Scanner == nil {
		return nil
	}
	err := g.Scanner.Scan(ctx, r)
	switch {
	case err == nil:
		return nil
	case IsInfected(err):
		g.logf("[avscan] SECURITY upload rejected: malware detected in %s: %v", label, err)
		return err
	case g.Mode == ModeOptional:
		g.logf("[avscan] WARNING scanner unavailable, accepting %s unscanned (CLAMAV_MODE=optional): %v", label, err)
		return nil
	default:
		g.logf("[avscan] upload refused, scanner unavailable for %s (CLAMAV_MODE=required): %v", label, err)
		if errors.Is(err, ErrUnavailable) {
			return err
		}
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
}

// FromEnv builds a Guard from CLAMAV_ADDRESS, CLAMAV_MODE and CLAMAV_TIMEOUT.
func FromEnv() *Guard {
	mode := Mode(strings.ToLower(strings.TrimSpace(os.Getenv("CLAMAV_MODE"))))
	switch mode {
	case ModeRequired, ModeOptional, ModeOff:
	case "":
		mode = ModeRequired
	default:
		log.Printf("[avscan] unknown CLAMAV_MODE %q, using %q", mode, ModeRequired)
		mode = ModeRequired
	}
	addr := strings.TrimSpace(os.Getenv("CLAMAV_ADDRESS"))
	if addr == "" {
		addr = DefaultAddress
	}
	timeout := DefaultTimeout
	if v := strings.TrimSpace(os.Getenv("CLAMAV_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			timeout = d
		}
	}
	return &Guard{Scanner: &Clamd{Address: addr, Timeout: timeout}, Mode: mode}
}

var (
	defaultOnce  sync.Once
	defaultGuard *Guard
	defaultMu    sync.RWMutex
)

// Default returns the process-wide guard, built from the environment on first use.
func Default() *Guard {
	defaultOnce.Do(func() {
		g := FromEnv()
		defaultMu.Lock()
		if defaultGuard == nil {
			defaultGuard = g
		}
		defaultMu.Unlock()
		addr := ""
		if c, ok := g.Scanner.(*Clamd); ok {
			addr = c.Address
		}
		log.Printf("[avscan] upload malware scanning: mode=%s address=%s", g.Mode, addr)
	})
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultGuard
}

// SetDefault replaces the process-wide guard (for tests) and returns a
// function that restores the previous one.
func SetDefault(g *Guard) (restore func()) {
	defaultOnce.Do(func() {})
	defaultMu.Lock()
	prev := defaultGuard
	defaultGuard = g
	defaultMu.Unlock()
	return func() {
		defaultMu.Lock()
		defaultGuard = prev
		defaultMu.Unlock()
	}
}
