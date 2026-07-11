package main

import (
	"bufio"
	"encoding/base64"
	"net"
	"testing"
)

func TestParseVlessUUID(t *testing.T) {
	valid := "12345678-1234-1234-1234-1234567890ab"
	b, err := parseVlessUUID(valid)
	if err != nil {
		t.Fatalf("parseVlessUUID(valid) error: %v", err)
	}
	if len(b) != 16 {
		t.Fatalf("parseVlessUUID length = %d, want 16", len(b))
	}
	if _, err := parseVlessUUID("short"); err == nil {
		t.Errorf("expected error for short UUID")
	}
	// 32 chars but not hex.
	if _, err := parseVlessUUID("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"); err == nil {
		t.Errorf("expected error for non-hex UUID")
	}
}

func TestParseVlessURI(t *testing.T) {
	uri := "vless://11111111-2222-3333-4444-555555555555@cdn.example.com:8443?type=ws&security=tls&path=%2Fwspath&host=front.example.com&sni=real.example.com#MyProfile"
	p, err := parseVlessURI(uri)
	if err != nil {
		t.Fatalf("parseVlessURI error: %v", err)
	}
	if p.TunnelType != "vless" {
		t.Errorf("TunnelType = %q, want vless", p.TunnelType)
	}
	if p.Name != "MyProfile" {
		t.Errorf("Name = %q, want MyProfile", p.Name)
	}
	if p.VlessUuid != "11111111-2222-3333-4444-555555555555" {
		t.Errorf("VlessUuid = %q", p.VlessUuid)
	}
	if p.VlessCdnPort != 8443 {
		t.Errorf("VlessCdnPort = %d, want 8443", p.VlessCdnPort)
	}
	if p.VlessWsPath != "/wspath" {
		t.Errorf("VlessWsPath = %q, want /wspath (percent-decoded)", p.VlessWsPath)
	}
	// host param overrides domain; sni differs from host so FakeSni is set.
	if p.Domain != "front.example.com" {
		t.Errorf("Domain = %q, want front.example.com", p.Domain)
	}
	if p.FakeSni != "real.example.com" {
		t.Errorf("FakeSni = %q, want real.example.com", p.FakeSni)
	}
}

func TestParseVlessURIDefaults(t *testing.T) {
	// No port, no params: port defaults to 443, path to "/", name to VLESS.
	p, err := parseVlessURI("vless://abcd@server.example.com")
	if err != nil {
		t.Fatalf("parseVlessURI error: %v", err)
	}
	if p.VlessCdnPort != 443 {
		t.Errorf("default CdnPort = %d, want 443", p.VlessCdnPort)
	}
	if p.VlessWsPath != "/" {
		t.Errorf("default WsPath = %q, want /", p.VlessWsPath)
	}
	if p.Name != "VLESS" {
		t.Errorf("default Name = %q, want VLESS", p.Name)
	}
	if p.FakeSni != "" {
		t.Errorf("FakeSni should be empty when sni==host, got %q", p.FakeSni)
	}
}

func TestParseVlessURIMissingUUID(t *testing.T) {
	if _, err := parseVlessURI("vless://server.example.com:443"); err == nil {
		t.Errorf("expected error when '@' (UUID) is missing")
	}
}

func TestBuildVlessHeader(t *testing.T) {
	uuid := make([]byte, 16)
	for i := range uuid {
		uuid[i] = byte(i)
	}

	t.Run("ipv4", func(t *testing.T) {
		h := buildVlessHeader(uuid, "1.2.3.4", 443)
		if h[0] != 0x00 {
			t.Errorf("version = %d, want 0", h[0])
		}
		if string(h[1:17]) != string(uuid) {
			t.Errorf("uuid mismatch")
		}
		if h[17] != 0x00 || h[18] != 0x01 {
			t.Errorf("addons/command = %d,%d want 0,1", h[17], h[18])
		}
		if h[19] != 0x01 || h[20] != 0xBB {
			t.Errorf("port bytes = %d,%d, want 1,187 (443)", h[19], h[20])
		}
		if h[21] != 0x01 {
			t.Errorf("atyp = %d, want 1 (IPv4)", h[21])
		}
		if h[22] != 1 || h[23] != 2 || h[24] != 3 || h[25] != 4 {
			t.Errorf("ipv4 bytes = %v", h[22:26])
		}
	})

	t.Run("domain", func(t *testing.T) {
		h := buildVlessHeader(uuid, "example.com", 80)
		if h[21] != 0x02 {
			t.Errorf("atyp = %d, want 2 (domain)", h[21])
		}
		if int(h[22]) != len("example.com") {
			t.Errorf("domain length = %d, want %d", h[22], len("example.com"))
		}
		if string(h[23:23+len("example.com")]) != "example.com" {
			t.Errorf("domain = %q", string(h[23:23+len("example.com")]))
		}
	})

	t.Run("ipv6", func(t *testing.T) {
		h := buildVlessHeader(uuid, "2001:db8::1", 443)
		if h[21] != 0x03 {
			t.Errorf("atyp = %d, want 3 (IPv6)", h[21])
		}
	})
}

func TestGenerateVlessWSKey(t *testing.T) {
	k := generateVlessWSKey()
	dec, err := base64.StdEncoding.DecodeString(k)
	if err != nil {
		t.Fatalf("key is not valid base64: %v", err)
	}
	if len(dec) != 16 {
		t.Errorf("decoded key length = %d, want 16", len(dec))
	}
	if generateVlessWSKey() == k {
		t.Errorf("keys should be random, got two identical values")
	}
}

func TestVlessWSFrameRoundTrip(t *testing.T) {
	payloads := [][]byte{
		[]byte("hi"),
		make([]byte, 200),   // 2-byte extended length
		make([]byte, 70000), // 8-byte extended length
	}
	for i := range payloads[1] {
		payloads[1][i] = byte(i)
	}

	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	go func() {
		for _, p := range payloads {
			_ = writeWSBinaryFrame(a, p)
		}
	}()

	br := bufio.NewReader(b)
	for i, want := range payloads {
		got, err := readWSFrame(br)
		if err != nil {
			t.Fatalf("readWSFrame(%d): %v", i, err)
		}
		if len(got) != len(want) {
			t.Fatalf("frame %d length = %d, want %d", i, len(got), len(want))
		}
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("frame %d byte %d = %d, want %d", i, j, got[j], want[j])
			}
		}
	}
}
