package upload

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"audit-service/pkg/avscan"

	"github.com/gin-gonic/gin"
)

var testPDF = []byte("%PDF-1.4\n1 0 obj<<>>endobj\n%%EOF")

// useRoot points UploadsRoot at a temp dir and disables scanning unless a test sets its own guard.
func useRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	old := UploadsRoot
	UploadsRoot = root
	t.Cleanup(func() { UploadsRoot = old })
	restore := avscan.SetDefault(&avscan.Guard{Mode: avscan.ModeOff})
	t.Cleanup(restore)
	return root
}

func TestCheckNameRejectsTraversal(t *testing.T) {
	for _, name := range []string{
		"../../etc/passwd",
		"..\\..\\win.ini",
		"a/../../b.pdf",
		"/abs/path.pdf",
		"..",
		".",
		"",
		"   ",
		"nul\x00.pdf",
	} {
		if _, e := CheckName(name); e == nil {
			t.Errorf("CheckName(%q) accepted a traversal or empty name", name)
		}
	}
}

func TestCheckNameKeepsSafeNames(t *testing.T) {
	got, e := CheckName("Laporan Audit (final).pdf")
	if e != nil || got != "Laporan Audit (final).pdf" {
		t.Fatalf("got %q, %v", got, e)
	}
}

func TestDisplayNameStripsDirectories(t *testing.T) {
	if got := DisplayName(`C:\Users\x\plan.pdf`); got != "plan.pdf" {
		t.Fatalf("got %q", got)
	}
	if got := DisplayName("../../x.pdf"); got != "x.pdf" {
		t.Fatalf("got %q", got)
	}
}

func TestStoredNameIsServerGenerated(t *testing.T) {
	a := StoredName("evil..pdf")
	b := StoredName("evil..pdf")
	if a == b {
		t.Fatal("stored names must be unique")
	}
	if strings.ContainsAny(a, `/\ `) || strings.Contains(a, "..") {
		t.Fatalf("unsafe stored name %q", a)
	}
	if !strings.HasSuffix(a, ".pdf") {
		t.Fatalf("extension lost: %q", a)
	}
}

func TestConfineFolder(t *testing.T) {
	ok := map[string]string{
		"":                         "Auditsphere",
		"Auditsphere/charter":      "Auditsphere/charter",
		"Auditsphere/fieldwork/ST": "Auditsphere/fieldwork/ST",
		"fieldwork/ST-1":           "Auditsphere/fieldwork/ST-1",
		"/Auditsphere/sop/":        "Auditsphere/sop",
	}
	for in, want := range ok {
		got, e := ConfineFolder(in)
		if e != nil || got != want {
			t.Errorf("ConfineFolder(%q) = %q, %v; want %q", in, got, e, want)
		}
	}
	for _, in := range []string{
		"Auditsphere/../../x",
		"../../etc",
		"..",
		"a/./..",
		`Auditsphere\..\..\x`,
		"a\x00b",
	} {
		if got, e := ConfineFolder(in); e == nil {
			t.Errorf("ConfineFolder(%q) accepted %q", in, got)
		}
	}
}

func TestDisallowedExtensionRejected(t *testing.T) {
	useRoot(t)
	for _, name := range []string{"payload.html", "run.exe", "page.svg", "script.js", "archive.zip", "noext"} {
		_, e := Validate(context.Background(), name, testPDF, DefaultMaxBytes, DocumentExtensions)
		if e == nil || e.Code != "FILE_TYPE_NOT_ALLOWED" {
			t.Errorf("%s: want FILE_TYPE_NOT_ALLOWED, got %+v", name, e)
		}
	}
}

func TestContentMustMatchExtension(t *testing.T) {
	// An HTML page renamed to .pdf, and a PNG renamed to .docx.
	if e := CheckContent(".pdf", []byte("<html><script>alert(1)</script></html>")); e == nil {
		t.Error("html bytes accepted as pdf")
	}
	png := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...)
	if e := CheckContent(".docx", png); e == nil {
		t.Error("png bytes accepted as docx")
	}
	if e := CheckContent(".pdf", testPDF); e != nil {
		t.Errorf("real pdf rejected: %v", e)
	}
	if e := CheckContent(".txt", []byte("<script>x</script>")); e == nil {
		t.Error("html accepted as txt")
	}
}

func TestOversizeRejected(t *testing.T) {
	useRoot(t)
	big := bytes.Repeat([]byte("a"), 2048)
	_, e := Validate(context.Background(), "big.txt", big, 1024, DocumentExtensions)
	if e == nil || e.Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("want 413, got %+v", e)
	}
}

// multipartRequest builds a multipart body with a single "file" part.
func multipartRequest(t *testing.T, filename string, data []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(data)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestBodyOverCapGivesRequestEntityTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const max = 4096
	r := gin.New()
	var status int
	r.POST("/upload", func(c *gin.Context) {
		if e := ParseForm(c, max); e != nil {
			status = e.Status
			return
		}
		_, e := FormFile(c, "file", max)
		if e != nil {
			status = e.Status
			return
		}
		status = http.StatusOK
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, multipartRequest(t, "huge.pdf", bytes.Repeat([]byte("x"), 5*max)))
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d, want 413", status)
	}
}

func TestStoreNeverWritesOutsideRootAndNeverOverwrites(t *testing.T) {
	root := useRoot(t)
	dir := filepath.Join(root, "Auditsphere", "fieldwork")
	n1, p1, err := Store(dir, "plan.pdf", testPDF)
	if err != nil {
		t.Fatal(err)
	}
	n2, p2, err := Store(dir, "plan.pdf", []byte("second"))
	if err != nil {
		t.Fatal(err)
	}
	if n1 == n2 || p1 == p2 {
		t.Fatal("second upload reused the first stored name")
	}
	got, _ := os.ReadFile(p1)
	if !bytes.Equal(got, testPDF) {
		t.Fatal("first file was overwritten")
	}
	if _, _, err := Store(filepath.Join(root, "..", "outside"), "x.pdf", testPDF); err == nil {
		t.Fatal("Store accepted a directory outside the uploads root")
	}
}

func TestResolveStoredRefusesOutsideRoot(t *testing.T) {
	root := useRoot(t)
	outside := filepath.Join(filepath.Dir(root), "secret.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outside)
	for _, p := range []string{outside, filepath.Join(root, "..", "secret.txt"), "../secret.txt"} {
		if _, err := ResolveStored(p); err == nil {
			t.Errorf("ResolveStored(%q) allowed a path outside the root", p)
		}
	}
	inside := filepath.Join(root, "ok.pdf")
	os.WriteFile(inside, testPDF, 0o600)
	if _, err := ResolveStored(inside); err != nil {
		t.Errorf("inside path refused: %v", err)
	}
}

func TestValidateWithEICARWhenScannerRequired(t *testing.T) {
	useRoot(t)
	// Fake clamd that reports every stream as infected.
	addr := startFakeInfectedClamd(t)
	restore := avscan.SetDefault(&avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	defer restore()
	eicar := []byte("X5O!P%@AP[4\\PZX54(P^)7CC)7}$" + "EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*")
	_, e := Validate(context.Background(), "eicar.txt", eicar, DefaultMaxBytes, DocumentExtensions)
	if e == nil || e.Code != "MALWARE_DETECTED" || e.Status != http.StatusUnprocessableEntity {
		t.Fatalf("want MALWARE_DETECTED 422, got %+v", e)
	}
}

func TestValidateFailsClosedWhenScannerDown(t *testing.T) {
	useRoot(t)
	ln, _ := netListenClosed(t)
	restore := avscan.SetDefault(&avscan.Guard{Scanner: &avscan.Clamd{Address: ln}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	defer restore()
	_, e := Validate(context.Background(), "plan.pdf", testPDF, DefaultMaxBytes, DocumentExtensions)
	if e == nil || e.Code != "SCANNER_UNAVAILABLE" || e.Status != http.StatusServiceUnavailable {
		t.Fatalf("want SCANNER_UNAVAILABLE 503, got %+v", e)
	}
}

func TestValidateAcceptsPDFWithScannerClean(t *testing.T) {
	useRoot(t)
	addr := startFakeCleanClamd(t)
	restore := avscan.SetDefault(&avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	defer restore()
	name, e := Validate(context.Background(), "Plan Tahunan.pdf", testPDF, DefaultMaxBytes, DocumentExtensions)
	if e != nil {
		t.Fatalf("clean pdf rejected: %+v", e)
	}
	if name != "Plan Tahunan.pdf" {
		t.Fatalf("display name %q", name)
	}
}

// fakeScanner starts a TCP listener that reads one INSTREAM request and
// answers with reply. It returns the address.
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

func startFakeInfectedClamd(t *testing.T) string {
	return fakeScanner(t, "stream: Eicar-Test-Signature FOUND\x00")
}

func startFakeCleanClamd(t *testing.T) string {
	return fakeScanner(t, "stream: OK\x00")
}

// netListenClosed returns an address where nothing listens.
func netListenClosed(t *testing.T) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr, func() {}
}
