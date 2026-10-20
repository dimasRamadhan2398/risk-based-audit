package middleware

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// redactedValue replaces the value of a sensitive query parameter in logs.
const redactedValue = "REDACTED"

// sensitiveQueryMarkers are matched case-insensitively against query parameter names.
var sensitiveQueryMarkers = []string{"token", "secret", "key", "password", "passwd", "signature", "auth"}

// RedactQuery returns rawQuery with the values of sensitive parameters replaced,
// so access logs never carry live tokens, secrets or passwords. A query that
// cannot be parsed is dropped rather than logged.
func RedactQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "unparseable-query-redacted"
	}
	for name, vals := range values {
		if isSensitiveQueryName(name) {
			for i := range vals {
				vals[i] = redactedValue
			}
		}
	}
	return values.Encode()
}

func isSensitiveQueryName(name string) bool {
	lower := strings.ToLower(name)
	for _, marker := range sensitiveQueryMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// RedactingLogger is gin's default access log with sensitive query values redacted.
func RedactingLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		path := p.Path
		if p.Request != nil && p.Request.URL.RawQuery != "" {
			path = path + "?" + RedactQuery(p.Request.URL.RawQuery)
		}
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %#v\n%s",
			p.TimeStamp.Format("2006/01/02 - 15:04:05"), p.StatusCode, p.Latency, p.ClientIP, p.Method, path, p.ErrorMessage)
	})
}
