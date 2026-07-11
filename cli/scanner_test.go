package main

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestTunnelTestResultScore(t *testing.T) {
	cases := []struct {
		name string
		r    TunnelTestResult
		want int
	}{
		{"all false", TunnelTestResult{}, 0},
		{"all true", TunnelTestResult{
			NSSupport: true, TXTSupport: true, RandomSub: true,
			TunnelRealism: true, EDNS0Support: true, NXDOMAINCorrect: true,
		}, 6},
		{"three true", TunnelTestResult{NSSupport: true, TXTSupport: true, EDNS0Support: true}, 3},
		{"EDNSMaxPayload does not count", TunnelTestResult{EDNSMaxPayload: 4096}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.r.Score(); got != tc.want {
				t.Errorf("Score() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTunnelTestResultDetails(t *testing.T) {
	r := TunnelTestResult{NSSupport: true, EDNS0Support: true, EDNSMaxPayload: 1232}
	got := r.Details()
	if !strings.Contains(got, "NS→A✓") {
		t.Errorf("Details() = %q, missing NS→A✓", got)
	}
	if !strings.Contains(got, "TXT✗") {
		t.Errorf("Details() = %q, missing TXT✗", got)
	}
	if !strings.Contains(got, "EDNS✓(1232)") {
		t.Errorf("Details() = %q, missing EDNS payload annotation", got)
	}
	// When EDNS0 is unsupported the payload size must not be shown.
	r2 := TunnelTestResult{EDNSMaxPayload: 1232}
	if strings.Contains(r2.Details(), "(1232)") {
		t.Errorf("Details() should not show payload when EDNS0 unsupported: %q", r2.Details())
	}
}

func TestEncodeDNSName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []byte
	}{
		{"simple", "a.bc", []byte{1, 'a', 2, 'b', 'c', 0}},
		{"root", "", []byte{0}},
		{"trailing dot", "x.", []byte{1, 'x', 0}},
		{"empty labels skipped", "a..b", []byte{1, 'a', 1, 'b', 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := encodeDNSName(tc.in)
			if string(got) != string(tc.want) {
				t.Errorf("encodeDNSName(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestEncodeDNSNameLongLabelTruncated(t *testing.T) {
	long := strings.Repeat("z", 100)
	got := encodeDNSName(long)
	if got[0] != 63 {
		t.Fatalf("expected label length capped at 63, got %d", got[0])
	}
	if len(got) != 1+63+1 {
		t.Fatalf("expected total length %d, got %d", 1+63+1, len(got))
	}
}

// buildDNSNameRoundtrip encodes a name and returns a packet slice where the
// name begins at offset 0 so decodeDNSName/skipDNSName can be exercised.
func TestDNSNameRoundtrip(t *testing.T) {
	for _, name := range []string{"example.com", "a.b.c.d", "single"} {
		pkt := encodeDNSName(name)
		if got := decodeDNSName(pkt, 0); got != name {
			t.Errorf("decodeDNSName(encode(%q)) = %q", name, got)
		}
		if got := skipDNSName(pkt, 0); got != len(pkt) {
			t.Errorf("skipDNSName(encode(%q)) = %d, want %d", name, got, len(pkt))
		}
	}
}

func TestSkipDNSNamePointer(t *testing.T) {
	// A compression pointer is two bytes and skipDNSName should advance by 2.
	pkt := []byte{0xC0, 0x0C}
	if got := skipDNSName(pkt, 0); got != 2 {
		t.Errorf("skipDNSName(pointer) = %d, want 2", got)
	}
}

func TestDecodeDNSNameFollowsPointer(t *testing.T) {
	// Layout: [0..] "com" root at offset 0; a name at offset 5 that is
	// label "www" then a pointer back to offset 0.
	pkt := []byte{3, 'c', 'o', 'm', 0, 3, 'w', 'w', 'w', 0xC0, 0x00}
	if got := decodeDNSName(pkt, 5); got != "www.com" {
		t.Errorf("decodeDNSName with pointer = %q, want www.com", got)
	}
}

func TestDecodeDNSNameLoopGuard(t *testing.T) {
	// A pointer that points to itself must not loop forever.
	pkt := []byte{0xC0, 0x00}
	_ = decodeDNSName(pkt, 0) // must return without hanging
}

func TestBuildEDNS0OPT(t *testing.T) {
	opt := buildEDNS0OPT(4096)
	if len(opt) != 11 {
		t.Fatalf("OPT length = %d, want 11", len(opt))
	}
	if opt[0] != 0x00 {
		t.Errorf("root name byte = %d, want 0", opt[0])
	}
	if binary.BigEndian.Uint16(opt[1:3]) != dnsTypeOPT {
		t.Errorf("type = %d, want %d", binary.BigEndian.Uint16(opt[1:3]), dnsTypeOPT)
	}
	if binary.BigEndian.Uint16(opt[3:5]) != 4096 {
		t.Errorf("payload size = %d, want 4096", binary.BigEndian.Uint16(opt[3:5]))
	}
}

// dnsHeader builds a 12-byte DNS header with the given section counts.
func dnsHeader(qd, an, ns, ar uint16) []byte {
	h := make([]byte, 12)
	binary.BigEndian.PutUint16(h[4:6], qd)
	binary.BigEndian.PutUint16(h[6:8], an)
	binary.BigEndian.PutUint16(h[8:10], ns)
	binary.BigEndian.PutUint16(h[10:12], ar)
	return h
}

func TestHasOPTRecord(t *testing.T) {
	// Header with 1 question and 1 additional (the OPT record).
	pkt := dnsHeader(1, 0, 0, 1)
	pkt = append(pkt, encodeDNSName("t.example.com")...) // question name
	pkt = append(pkt, 0x00, 0x10, 0x00, 0x01)            // QTYPE TXT, QCLASS IN
	pkt = append(pkt, buildEDNS0OPT(1232)...)            // additional OPT
	if !hasOPTRecord(pkt) {
		t.Errorf("hasOPTRecord() = false, want true")
	}
}

func TestHasOPTRecordNone(t *testing.T) {
	if hasOPTRecord(dnsHeader(0, 0, 0, 0)) {
		t.Errorf("hasOPTRecord(empty) = true, want false")
	}
	if hasOPTRecord([]byte{0, 1, 2}) {
		t.Errorf("hasOPTRecord(short) = true, want false")
	}
}

func TestExtractNSHost(t *testing.T) {
	nsName := encodeDNSName("ns1.example.com")
	pkt := dnsHeader(1, 1, 0, 0)
	pkt = append(pkt, encodeDNSName("example.com")...) // question
	pkt = append(pkt, 0x00, byte(dnsTypeNS), 0x00, 0x01)
	// answer: name, type NS, class, ttl(4), rdlen, rdata(nsName)
	pkt = append(pkt, encodeDNSName("example.com")...)
	pkt = append(pkt, 0x00, byte(dnsTypeNS)) // type
	pkt = append(pkt, 0x00, 0x01)            // class
	pkt = append(pkt, 0, 0, 0, 0)            // ttl
	rdlen := make([]byte, 2)
	binary.BigEndian.PutUint16(rdlen, uint16(len(nsName)))
	pkt = append(pkt, rdlen...)
	pkt = append(pkt, nsName...)

	if got := extractNSHost(pkt); got != "ns1.example.com" {
		t.Errorf("extractNSHost() = %q, want ns1.example.com", got)
	}
}

func TestExtractNSHostEmpty(t *testing.T) {
	if got := extractNSHost([]byte{0, 1}); got != "" {
		t.Errorf("extractNSHost(short) = %q, want empty", got)
	}
}

// buildTXTResponse constructs a minimal DNS response carrying one TXT record.
func buildTXTResponse(strings_ ...string) []byte {
	pkt := dnsHeader(1, 1, 0, 0)
	pkt = append(pkt, encodeDNSName("t.example.com")...)
	pkt = append(pkt, 0x00, byte(dnsTypeTXT), 0x00, 0x01)
	pkt = append(pkt, encodeDNSName("t.example.com")...)
	pkt = append(pkt, 0x00, byte(dnsTypeTXT), 0x00, 0x01, 0, 0, 0, 0)
	var rdata []byte
	for _, s := range strings_ {
		rdata = append(rdata, byte(len(s)))
		rdata = append(rdata, []byte(s)...)
	}
	rdlen := make([]byte, 2)
	binary.BigEndian.PutUint16(rdlen, uint16(len(rdata)))
	pkt = append(pkt, rdlen...)
	pkt = append(pkt, rdata...)
	return pkt
}

func TestExtractTXTData(t *testing.T) {
	pkt := buildTXTResponse("hello", "world")
	if got := extractTXTData(pkt); got != "helloworld" {
		t.Errorf("extractTXTData() = %q, want helloworld", got)
	}
	raw := extractTXTRawData(pkt)
	if string(raw) != "helloworld" {
		t.Errorf("extractTXTRawData() = %q, want helloworld", string(raw))
	}
}

func TestExtractTXTDataEmpty(t *testing.T) {
	if got := extractTXTData(dnsHeader(0, 0, 0, 0)); got != "" {
		t.Errorf("extractTXTData(no answers) = %q, want empty", got)
	}
	if raw := extractTXTRawData([]byte{1, 2, 3}); raw != nil {
		t.Errorf("extractTXTRawData(short) = %v, want nil", raw)
	}
}

func TestGetParentDomain(t *testing.T) {
	cases := map[string]string{
		"t.example.com":     "example.com",
		"a.b.example.co.uk": "b.example.co.uk",
		"example.com":       "example.com", // no grandparent
		"localhost":         "localhost",
	}
	for in, want := range cases {
		if got := getParentDomain(in); got != want {
			t.Errorf("getParentDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRandomLabel(t *testing.T) {
	const allowed = "abcdefghijklmnopqrstuvwxyz0123456789"
	l := randomLabel(20)
	if len(l) != 20 {
		t.Fatalf("randomLabel(20) length = %d", len(l))
	}
	for _, c := range l {
		if !strings.ContainsRune(allowed, c) {
			t.Fatalf("randomLabel produced disallowed char %q", c)
		}
	}
	if randomLabel(0) != "" {
		t.Errorf("randomLabel(0) should be empty")
	}
}

func TestBase32Encode(t *testing.T) {
	// Known RFC 4648 vectors using the uppercase alphabet used by base32Encode.
	cases := map[string]string{
		"":      "",
		"f":     "MY",
		"fo":    "MZXQ",
		"foo":   "MZXW6",
		"foob":  "MZXW6YQ",
		"fooba": "MZXW6YTB",
	}
	for in, want := range cases {
		if got := base32Encode([]byte(in)); got != want {
			t.Errorf("base32Encode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitLabels(t *testing.T) {
	got := splitLabels("aaabbbcc", 3)
	want := []string{"aaa", "bbb", "cc"}
	if len(got) != len(want) {
		t.Fatalf("splitLabels len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("splitLabels[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if len(splitLabels("", 3)) != 0 {
		t.Errorf("splitLabels(empty) should be empty")
	}
	if l := splitLabels("ab", 3); len(l) != 1 || l[0] != "ab" {
		t.Errorf("splitLabels short = %v, want [ab]", l)
	}
}

type fakeTimeoutErr struct{ timeout bool }

func (e fakeTimeoutErr) Error() string   { return "fake" }
func (e fakeTimeoutErr) Timeout() bool   { return e.timeout }
func (e fakeTimeoutErr) Temporary() bool { return false }

func TestIsTimeout(t *testing.T) {
	if isTimeout(nil) {
		t.Errorf("isTimeout(nil) = true, want false")
	}
	if isTimeout(errors.New("plain")) {
		t.Errorf("isTimeout(plain error) = true, want false")
	}
	var netErr net.Error = fakeTimeoutErr{timeout: true}
	if !isTimeout(netErr) {
		t.Errorf("isTimeout(timeout net.Error) = false, want true")
	}
	if isTimeout(fakeTimeoutErr{timeout: false}) {
		t.Errorf("isTimeout(non-timeout net.Error) = true, want false")
	}
}

func TestLoadIPList(t *testing.T) {
	content := strings.Join([]string{
		"# comment",
		"1.1.1.1",
		"  8.8.8.8:53  ",
		"1.1.1.1",           // duplicate
		"not-an-ip",
		"",
		"9.9.9.9 extra col",
		"2001:4860:4860::8888",
	}, "\n")
	got := LoadIPList(content)
	want := []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}
	if len(got) != len(want) {
		t.Fatalf("LoadIPList len = %d (%v), want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("LoadIPList[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
