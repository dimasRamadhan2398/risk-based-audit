package activitylog

import (
	"net"
	"strings"
)

// Redacted replaces the value of a sensitive detail key.
const Redacted = "[REDACTED]"

// sensitiveKeyMarkers are matched case-insensitively against detail keys.
// "code" covers MFA/OTP codes, "jti" the step-up token id.
var sensitiveKeyMarkers = []string{
	"password", "passwd", "pwd", "secret", "token", "jti", "otp", "code",
	"authorization", "cookie", "api_key", "apikey", "private", "signature", "credential",
}

// IsSensitiveKey reports whether a detail key must never be stored in clear.
func IsSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, marker := range sensitiveKeyMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// Redact returns a copy of details with sensitive values replaced, recursing
// into nested objects and arrays. Nil stays nil.
func Redact(details map[string]interface{}) map[string]interface{} {
	if details == nil {
		return nil
	}
	out := make(map[string]interface{}, len(details))
	for k, v := range details {
		if IsSensitiveKey(k) {
			out[k] = Redacted
			continue
		}
		out[k] = redactValue(v)
	}
	return out
}

func redactValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return Redact(t)
	case map[string]string:
		m := make(map[string]interface{}, len(t))
		for k, s := range t {
			m[k] = s
		}
		return Redact(m)
	case []interface{}:
		out := make([]interface{}, len(t))
		for i := range t {
			out[i] = redactValue(t[i])
		}
		return out
	default:
		return v
	}
}

// MaskIP hides the host part of an address: IPv4 keeps the first two octets
// ("192.168.x.x"), IPv6 the first two groups ("2001:db8:x:x:x:x:x:x").
// Anything that is not an IP is replaced entirely.
func MaskIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "x.x.x.x"
	}
	if v4 := parsed.To4(); v4 != nil {
		parts := strings.Split(v4.String(), ".")
		return parts[0] + "." + parts[1] + ".x.x"
	}
	expanded := parsed.To16()
	groups := []string{}
	for i := 0; i < 4; i += 2 {
		groups = append(groups, strings.TrimLeft(hexByte(expanded[i])+hexByte(expanded[i+1]), "0"))
	}
	for i := range groups {
		if groups[i] == "" {
			groups[i] = "0"
		}
	}
	return groups[0] + ":" + groups[1] + ":x:x:x:x:x:x"
}

func hexByte(b byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[b>>4], digits[b&0x0f]})
}
