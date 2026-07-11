package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

func encodeSlipnet(fields []string) string {
	raw := strings.Join(fields, "|")
	return "slipnet://" + base64.StdEncoding.EncodeToString([]byte(raw))
}

func TestParseURIBasic(t *testing.T) {
	fields := []string{
		"v18",         // 0 version
		"dnstt",       // 1 tunnel type
		"My Server",   // 2 name
		"t.example.com", // 3 domain
		"1.1.1.1,8.8.8.8", // 4 resolvers
		"1",           // 5 auth mode
		"30",          // 6 keepalive
		"US",          // 7 cc
		"9050",        // 8 port
		"0.0.0.0",     // 9 host
		"1",           // 10 gso
		"deadbeef",    // 11 public key
	}
	p, err := parseURI(encodeSlipnet(fields))
	if err != nil {
		t.Fatalf("parseURI error: %v", err)
	}
	if p.Version != "v18" || p.TunnelType != "dnstt" || p.Name != "My Server" {
		t.Errorf("basic fields mismatch: %+v", p)
	}
	if p.Domain != "t.example.com" || p.Resolvers != "1.1.1.1,8.8.8.8" {
		t.Errorf("domain/resolvers mismatch: %+v", p)
	}
	if !p.AuthMode {
		t.Errorf("AuthMode = false, want true")
	}
	if p.KeepAlive != 30 {
		t.Errorf("KeepAlive = %d, want 30", p.KeepAlive)
	}
	if p.CC != "US" {
		t.Errorf("CC = %q, want US", p.CC)
	}
	if p.Port != 9050 {
		t.Errorf("Port = %d, want 9050", p.Port)
	}
	if p.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want 0.0.0.0", p.Host)
	}
	if !p.GSO {
		t.Errorf("GSO = false, want true")
	}
	if p.PublicKey != "deadbeef" {
		t.Errorf("PublicKey = %q, want deadbeef", p.PublicKey)
	}
}

func TestParseURIDefaultsWhenFieldsBlank(t *testing.T) {
	// Empty host/port fields should fall back to defaults (127.0.0.1:10880).
	fields := []string{"v18", "dnstt", "n", "d", "r", "0", "0", "", "", "", "0", "pk"}
	p, err := parseURI(encodeSlipnet(fields))
	if err != nil {
		t.Fatalf("parseURI error: %v", err)
	}
	if p.Host != "127.0.0.1" {
		t.Errorf("default Host = %q, want 127.0.0.1", p.Host)
	}
	if p.Port != 10880 {
		t.Errorf("default Port = %d, want 10880", p.Port)
	}
	if p.AuthMode || p.GSO {
		t.Errorf("AuthMode/GSO should be false")
	}
}

func TestParseURIInvalidScheme(t *testing.T) {
	if _, err := parseURI("http://example.com"); err == nil {
		t.Errorf("expected error for invalid scheme")
	}
}

func TestParseURIBadBase64(t *testing.T) {
	if _, err := parseURI("slipnet://!!!not base64!!!@@@"); err == nil {
		t.Errorf("expected base64 decode error")
	}
}

func TestParseURITooFewFields(t *testing.T) {
	if _, err := parseURI(encodeSlipnet([]string{"a", "b", "c"})); err == nil {
		t.Errorf("expected error for too few fields")
	}
}

func TestParseURIWhitespaceStripped(t *testing.T) {
	fields := []string{"v18", "dnstt", "n", "d", "r", "0", "0", "c", "0", "", "0", "pk"}
	uri := encodeSlipnet(fields)
	// Inject line wrapping into the base64 payload.
	body := strings.TrimPrefix(uri, "slipnet://")
	wrapped := "slipnet://" + body[:8] + "\n  " + body[8:]
	if _, err := parseURI(wrapped); err != nil {
		t.Errorf("parseURI should strip whitespace/newlines, got: %v", err)
	}
}

func TestParseURIDispatchesVless(t *testing.T) {
	p, err := parseURI("vless://abcd@server.example.com:443")
	if err != nil {
		t.Fatalf("parseURI(vless) error: %v", err)
	}
	if p.TunnelType != "vless" {
		t.Errorf("TunnelType = %q, want vless", p.TunnelType)
	}
}
