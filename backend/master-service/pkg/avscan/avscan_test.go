package avscan

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// eicar is the standard antivirus test string, assembled at run time so the
// repository itself does not contain a detectable file.
func eicar() []byte {
	return []byte("X5O!P%@AP[4\\PZX54(P^)7CC)7}$" + "EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*")
}

// fakeClamd answers one INSTREAM request with reply and records the bytes it received.
func fakeClamd(t *testing.T, reply string) (addr string, received func() []byte) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	var got []byte
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		cmd, _ := r.ReadString(0)
		if cmd != "zINSTREAM\x00" {
			return
		}
		for {
			var n uint32
			if err := binary.Read(r, binary.BigEndian, &n); err != nil {
				return
			}
			if n == 0 {
				break
			}
			buf := make([]byte, n)
			if _, err := io.ReadFull(r, buf); err != nil {
				return
			}
			got = append(got, buf...)
		}
		_, _ = conn.Write([]byte(reply))
	}()
	t.Cleanup(func() { <-done })
	return ln.Addr().String(), func() []byte { return got }
}

func TestParseReply(t *testing.T) {
	cases := []struct {
		reply    string
		wantOK   bool
		wantSig  string
		wantErrU bool // ErrUnavailable expected
	}{
		{"stream: OK\x00", true, "", false},
		{"stream: Eicar-Signature FOUND\x00", false, "Eicar-Signature", false},
		{"stream: Win.Test.EICAR_HDB-1 FOUND\x00", false, "Win.Test.EICAR_HDB-1", false},
		{"INSTREAM size limit exceeded. ERROR\x00", false, "", true},
		{"", false, "", true},
	}
	for _, tc := range cases {
		err := parseReply(tc.reply)
		switch {
		case tc.wantOK:
			if err != nil {
				t.Errorf("%q: want nil, got %v", tc.reply, err)
			}
		case tc.wantErrU:
			if !errors.Is(err, ErrUnavailable) {
				t.Errorf("%q: want ErrUnavailable, got %v", tc.reply, err)
			}
		default:
			var ie *InfectedError
			if !errors.As(err, &ie) || ie.Signature != tc.wantSig {
				t.Errorf("%q: want infected %q, got %v", tc.reply, tc.wantSig, err)
			}
		}
	}
}

func TestClamdScanSendsStreamAndReadsVerdict(t *testing.T) {
	addr, received := fakeClamd(t, "stream: OK\x00")
	c := &Clamd{Address: addr, Timeout: 5 * time.Second}
	payload := bytes.Repeat([]byte("abc"), 100000) // spans several 64 KiB chunks
	if err := c.Scan(context.Background(), bytes.NewReader(payload)); err != nil {
		t.Fatalf("clean scan: %v", err)
	}
	if !bytes.Equal(received(), payload) {
		t.Fatalf("daemon received %d bytes, want %d identical bytes", len(received()), len(payload))
	}
}

func TestClamdScanReportsInfected(t *testing.T) {
	addr, _ := fakeClamd(t, "stream: Eicar-Test-Signature FOUND\x00")
	c := &Clamd{Address: addr, Timeout: 5 * time.Second}
	err := c.Scan(context.Background(), strings.NewReader("anything"))
	if !IsInfected(err) {
		t.Fatalf("want infected, got %v", err)
	}
}

func unreachableAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // nothing listens here now
	return addr
}

func TestGuardFailsClosedWhenRequired(t *testing.T) {
	g := &Guard{Scanner: &Clamd{Address: unreachableAddr(t), Timeout: time.Second}, Mode: ModeRequired, Logf: func(string, ...any) {}}
	err := g.CheckBytes(context.Background(), "x.pdf", []byte("%PDF-1.4"))
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("required mode with clamd down must refuse; got %v", err)
	}
}

func TestGuardFailsOpenWhenOptional(t *testing.T) {
	g := &Guard{Scanner: &Clamd{Address: unreachableAddr(t), Timeout: time.Second}, Mode: ModeOptional, Logf: func(string, ...any) {}}
	if err := g.CheckBytes(context.Background(), "x.pdf", []byte("%PDF-1.4")); err != nil {
		t.Fatalf("optional mode must accept when clamd is down; got %v", err)
	}
}

func TestGuardRefusesInfectedInEveryMode(t *testing.T) {
	for _, mode := range []Mode{ModeRequired, ModeOptional} {
		addr, _ := fakeClamd(t, "stream: Eicar-Test-Signature FOUND\x00")
		g := &Guard{Scanner: &Clamd{Address: addr, Timeout: 5 * time.Second}, Mode: mode, Logf: func(string, ...any) {}}
		if err := g.CheckBytes(context.Background(), "x.pdf", []byte("data")); !IsInfected(err) {
			t.Fatalf("mode %s: infected file was not refused: %v", mode, err)
		}
	}
}

func TestGuardOffSkipsScanner(t *testing.T) {
	g := &Guard{Scanner: &Clamd{Address: unreachableAddr(t)}, Mode: ModeOff}
	if err := g.CheckBytes(context.Background(), "x.pdf", []byte("x")); err != nil {
		t.Fatalf("off mode must not scan: %v", err)
	}
}

// TestLiveEICAR scans the EICAR test string with a real clamd. It runs only
// when CLAMAV_TEST_ADDR is set (for example 127.0.0.1:13310).
func TestLiveEICAR(t *testing.T) {
	addr := os.Getenv("CLAMAV_TEST_ADDR")
	if addr == "" {
		t.Skip("CLAMAV_TEST_ADDR not set; skipping live clamd test")
	}
	c := &Clamd{Address: addr, Timeout: 60 * time.Second}
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
	g := &Guard{Scanner: c, Mode: ModeRequired, Logf: func(string, ...any) {}}
	err := g.CheckBytes(context.Background(), "eicar.txt", eicar())
	if !IsInfected(err) {
		t.Fatalf("EICAR not detected: %v", err)
	}
	if err := g.CheckBytes(context.Background(), "clean.txt", []byte("hello, audit evidence\n")); err != nil {
		t.Fatalf("clean file rejected: %v", err)
	}
}
