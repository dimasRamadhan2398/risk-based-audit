package provisioner

import (
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		kind     LineKind
		key      string
		index    int
		status   string
		contains string // for log/malformed text
	}{
		{"progress running", `@@PROGRESS {"key":"build","index":3,"total":9,"status":"running","message":"Building 6 images"}`, LineProgress, "build", 3, "running", ""},
		{"progress done with CRLF", "@@PROGRESS {\"key\":\"build\",\"index\":3,\"total\":9,\"status\":\"done\",\"message\":\"ok\"}\r", LineProgress, "build", 3, "done", ""},
		{"progress skipped", `@@PROGRESS {"key":"email","index":6,"total":9,"status":"skipped","message":"local mode"}`, LineProgress, "email", 6, "skipped", ""},
		{"progress key wins over wrong index", `@@PROGRESS {"key":"ingress","index":2,"status":"failed","message":"x"}`, LineProgress, "ingress", 9, "failed", ""},
		{"progress index fallback", `@@PROGRESS {"index":4,"status":"running"}`, LineProgress, "migrate", 4, "running", ""},
		{"progress unknown status", `@@PROGRESS {"key":"build","index":3,"status":"pending"}`, LineMalformed, "", 0, "", "unknown status"},
		{"progress unknown step", `@@PROGRESS {"key":"nope","index":42,"status":"done"}`, LineMalformed, "", 0, "", "unknown step"},
		{"progress bad json", `@@PROGRESS {not json`, LineMalformed, "", 0, "", "malformed @@PROGRESS"},
		{"plain log", "==> [3/8] Building tenant service images...", LineLog, "", 0, "", "Building tenant"},
		{"prefix not at start is a log line", `echo @@PROGRESS {"key":"build"}`, LineLog, "", 0, "", "echo @@PROGRESS"},
		{"prefix without space is a log line", `@@PROGRESS{"key":"build"}`, LineLog, "", 0, "", "@@PROGRESS{"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ParseLine(tc.line)
			if p.Kind != tc.kind {
				t.Fatalf("kind = %v, want %v (%+v)", p.Kind, tc.kind, p)
			}
			if tc.kind == LineProgress {
				if p.Progress.Key != tc.key || p.Progress.Index != tc.index || p.Progress.Status != tc.status || p.Progress.Total != 9 {
					t.Fatalf("progress = %+v", p.Progress)
				}
			}
			if tc.contains != "" && !strings.Contains(p.Text, tc.contains) {
				t.Fatalf("text %q does not contain %q", p.Text, tc.contains)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	p := ParseLine(`@@RESULT {"slug":"bai","frontend_url":"https://bai.example","api_url":"https://api-bai.example","admin_username":"admin","admin_email":"cae@bai.co.id","admin_initial_password":"pw-123"}`)
	if p.Kind != LineResult {
		t.Fatalf("kind = %v", p.Kind)
	}
	r := p.Result
	if r.Slug != "bai" || r.FrontendURL != "https://bai.example" || r.APIURL != "https://api-bai.example" ||
		r.AdminUsername != "admin" || r.AdminEmail != "cae@bai.co.id" || r.AdminInitialPassword != "pw-123" {
		t.Fatalf("result = %+v", r)
	}
}

func TestMalformedResultHidesContent(t *testing.T) {
	p := ParseLine(`@@RESULT {"admin_initial_password":"leak-me"`)
	if p.Kind != LineMalformed {
		t.Fatalf("kind = %v", p.Kind)
	}
	if strings.Contains(p.Text, "leak-me") {
		t.Fatalf("malformed result text leaks the password: %q", p.Text)
	}
}

func TestParseResultNulls(t *testing.T) {
	p := ParseLine(`@@RESULT {"slug":"bai","frontend_url":"https://bai.example","api_url":"https://api-bai.example","admin_username":"admin","admin_email":null,"admin_initial_password":null}`)
	if p.Kind != LineResult || p.Result.AdminEmail != "" || p.Result.AdminInitialPassword != "" {
		t.Fatalf("parsed = %+v %+v", p, p.Result)
	}
}
