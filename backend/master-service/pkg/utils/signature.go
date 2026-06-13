package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
)

func generateSignature() string {
	return ""
}

func GenerateFingerprint(ipAddress, userAgent string) string {
	subnet := extractSubnet(ipAddress) // "192.168.1" instead of "192.168.1.105"
    data := fmt.Sprintf("%s|%s", subnet, userAgent)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func extractSubnet(ip string) string {
    parsed := net.ParseIP(ip)
    if parsed == nil {
        return ip
    }
    // Use /24 for IPv4 (drops last octet)
    mask := net.CIDRMask(24, 32)
    return parsed.Mask(mask).String()
}