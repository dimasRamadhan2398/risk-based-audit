package controllers

import "testing"

// Upload types are client-supplied. The extension decides what is served, and
// active content (HTML, SVG, XML, script) is never sent with its own type.
func TestGetDocumentContentTypeIgnoresClientType(t *testing.T) {
	cases := []struct {
		name, clientType, want string
	}{
		{"plan.pdf", "", "application/pdf"},
		{"plan.pdf", "text/html", "application/pdf"},
		{"page.html", "text/html", "application/octet-stream"},
		{"icon.svg", "image/svg+xml", "application/octet-stream"},
		{"data.xml", "application/xml", "application/octet-stream"},
		{"script.js", "text/javascript", "application/octet-stream"},
		{"noext", "text/html", "application/octet-stream"},
	}
	for _, tc := range cases {
		if got := getDocumentContentType(tc.name, tc.clientType); got != tc.want {
			t.Errorf("getDocumentContentType(%q, %q) = %q, want %q", tc.name, tc.clientType, got, tc.want)
		}
	}
}
