package upload

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"master-service/pkg/avscan"
)

var pdfBody = []byte("%PDF-1.4\n1 0 obj<<>>endobj\n%%EOF")

// useRoot points Root at a temp dir and disables scanning unless a test sets its own guard.
func useRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	old := Root
	Root = root
	t.Cleanup(func() { Root = old })
	restore := avscan.SetDefault(&avscan.Guard{Mode: avscan.ModeOff})
	t.Cleanup(restore)
	return root
}

// fakeScanner answers every INSTREAM request with reply.
func fakeScanner(t *testing.T, reply string) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				r := bufio.NewReader(conn)
				if _, err := r.ReadString(0); err != nil {
					return
				}
				for {
					var n uint32
					if err := binary.Read(r, binary.BigEndian, &n); err != nil || n == 0 {
						break
					}
					if _, err := io.CopyN(io.Discard, r, int64(n)); err != nil {
						return
					}
				}
				conn.Write([]byte(reply))
			}(conn)
		}
	}()
	return ln.Addr().String()
}

func TestCheckNameRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../../x.pdf", `a\b.pdf`, "a/b.pdf", "..", ".", "", "  ", "x\x00.pdf"} {
		if _, e := CheckName(name); e == nil {
			t.Errorf("CheckName(%q) accepted", name)
		}
	}
	if got, e := CheckName("Laporan QAR.pdf"); e != nil || got != "Laporan QAR.pdf" {
		t.Errorf("safe name: %q, %v", got, e)
	}
}

func TestExtensionAllowlist(t *testing.T) {
	for _, name := range []string{"page.html", "run.exe", "x.svg", "x.zip", "noext"} {
		if _, e := ExtensionOf(name); e == nil || e.Status != http.StatusUnsupportedMediaType {
			t.Errorf("%s: want 415, got %+v", name, e)
		}
	}
	if ext, e := ExtensionOf("Plan.PDF"); e != nil || ext != ".pdf" {
		t.Errorf("pdf: %q %+v", ext, e)
	}
}

func TestContentMismatchRejected(t *testing.T) {
	if e := CheckContent(".pdf", []byte("<html><script>x</script></html>")); e == nil {
		t.Error("html accepted as pdf")
	}
	if e := CheckContent(".txt", []byte("<script>alert(1)</script>")); e == nil {
		t.Error("script accepted as txt")
	}
	if e := CheckContent(".pdf", pdfBody); e != nil {
		t.Errorf("real pdf rejected: %+v", e)
	}
}

func TestValidateOversizeAndEmpty(t *testing.T) {
	useRoot(t)
	big := bytes.Repeat([]byte("%PDF-"), int(MaxBytes/5)+10)
	if _, e := Validate(context.Background(), "big.pdf", big); e == nil || e.Status != http.StatusRequestEntityTooLarge {
		t.Errorf("oversize: %+v", e)
	}
	if _, e := Validate(context.Background(), "empty.pdf", nil); e == nil || e.Status != http.StatusBadRequest {
		t.Errorf("empty: %+v", e)
	}
}

func TestValidateEICARRejectedWhenScannerSaysInfected(t *testing.T) {
	useRoot(t)
	addr := fakeScanner(t, "stream: Eicar-Test-Signature FOUND\x00")
	restore := avscan.SetDefault(&avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	defer restore()
	eicar := []byte("X5O!P%@AP[4\\PZX54(P^)7CC)7}$" + "EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*")
	_, e := Validate(context.Background(), "eicar.txt", eicar)
	if e == nil || e.Status != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %+v", e)
	}
}

func TestValidateFailsClosedWhenScannerDown(t *testing.T) {
	useRoot(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	restore := avscan.SetDefault(&avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	defer restore()
	if _, e := Validate(context.Background(), "plan.pdf", pdfBody); e == nil || e.Status != http.StatusServiceUnavailable {
		t.Fatalf("want 503 fail-closed, got %+v", e)
	}
}

func TestValidateAcceptsCleanPDF(t *testing.T) {
	useRoot(t)
	addr := fakeScanner(t, "stream: OK\x00")
	restore := avscan.SetDefault(&avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	defer restore()
	name, e := Validate(context.Background(), "QAR Q3.pdf", pdfBody)
	if e != nil || name != "QAR Q3.pdf" {
		t.Fatalf("clean pdf: %q %+v", name, e)
	}
}

func TestStoreConfinedAndNoOverwrite(t *testing.T) {
	root := useRoot(t)
	p1, err := Store("consulting", "plan.pdf", pdfBody)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := Store("consulting", "plan.pdf", []byte("second"))
	if err != nil {
		t.Fatal(err)
	}
	if p1 == p2 {
		t.Fatal("stored name reused")
	}
	if got, _ := os.ReadFile(p1); !bytes.Equal(got, pdfBody) {
		t.Fatal("first file overwritten")
	}
	if !strings.HasPrefix(p1, root) || !strings.HasSuffix(p1, ".pdf") {
		t.Fatalf("unexpected path %q", p1)
	}
	if _, err := Store("../../escape", "x.pdf", pdfBody); err == nil {
		t.Fatal("Store accepted a subdir outside the root")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "escape")); err == nil {
		t.Fatal("escape directory created")
	}
}

func TestLimitRejectsOversizeBody(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bytes.Repeat([]byte("a"), int(MaxBytes)*2)))
	Limit(rec, req)
	_, err := io.ReadAll(req.Body)
	if !IsTooLarge(err) {
		t.Fatalf("expected MaxBytesError, got %v", err)
	}
}
