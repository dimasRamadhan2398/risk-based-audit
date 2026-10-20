package middleware

import "testing"

func TestRedactQuery(t *testing.T) {
	cases := map[string]string{
		"":                             "",
		"page=2&size=10":               "page=2&size=10",
		"token=abc.def.ghi":            "token=REDACTED",
		"access_token=x&page=1":        "access_token=REDACTED&page=1",
		"api_key=k1&download=true":     "api_key=REDACTED&download=true",
		"DB_PASSWORD=p&secret=s":       "DB_PASSWORD=REDACTED&secret=REDACTED",
		"signature=sig&file=a%2Fb.pdf": "file=a%2Fb.pdf&signature=REDACTED",
	}
	for in, want := range cases {
		if got := RedactQuery(in); got != want {
			t.Errorf("RedactQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRedactQueryDropsUnparseable(t *testing.T) {
	if got := RedactQuery("%zz=1"); got != "unparseable-query-redacted" {
		t.Fatalf("unparseable query must not be logged verbatim, got %q", got)
	}
}
