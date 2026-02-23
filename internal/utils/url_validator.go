package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var (
	// ErrInvalidURL is returned when URL validation fails
	ErrInvalidURL = errors.New("invalid or forbidden URL")

	// Blocked private IP ranges (RFC 1918, loopback, link-local)
	privateIPBlocks = []*net.IPNet{
		// Loopback (127.0.0.0/8)
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		// Private networks (10.0.0.0/8)
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		// Private networks (172.16.0.0/12)
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.IPv4Mask(255, 240, 0, 0)},
		// Private networks (192.168.0.0/16)
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
		// Link-local (169.254.0.0/16) - includes AWS metadata
		{IP: net.IPv4(169, 254, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
	}
)

// ValidateExternalURL validates URL to prevent SSRF attacks
func ValidateExternalURL(urlStr string) error {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow HTTP and HTTPS schemes
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("forbidden URL scheme: %s (only http/https allowed)", scheme)
	}

	// Get hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return errors.New("URL must have a hostname")
	}

	// Block localhost variations
	lowercaseHost := strings.ToLower(hostname)
	forbiddenHosts := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"::1",
		"[::1]",
		// AWS metadata endpoints
		"169.254.169.254",
		"metadata.google.internal",
		"metadata",
		// Azure metadata
		"168.63.129.16",
	}

	for _, forbidden := range forbiddenHosts {
		if lowercaseHost == forbidden || strings.HasSuffix(lowercaseHost, "."+forbidden) {
			return fmt.Errorf("forbidden hostname: %s", hostname)
		}
	}

	// Resolve hostname to IP and check for private ranges
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("hostname %s does not resolve to any IP", hostname)
	}

	for _, ip := range ips {
		// Check if IP is in private range
		for _, block := range privateIPBlocks {
			if block.Contains(ip) {
				return fmt.Errorf("private IP address not allowed: %s resolves to %s", hostname, ip.String())
			}
		}
	}

	return nil
}

// isPrivateIP checks if an IP address is in a private range
func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
