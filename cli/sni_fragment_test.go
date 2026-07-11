package main

import (
	"bytes"
	"io"
	"net"
	"testing"
)

func TestSplitHalf(t *testing.T) {
	chunks := splitHalf([]byte("abcdef"))
	if len(chunks) != 2 {
		t.Fatalf("splitHalf returned %d chunks, want 2", len(chunks))
	}
	if string(chunks[0]) != "abc" || string(chunks[1]) != "def" {
		t.Errorf("splitHalf = %q,%q", chunks[0], chunks[1])
	}
}

func TestSplitMulti(t *testing.T) {
	chunks := splitMulti([]byte("aaabbbcc"), 3)
	if len(chunks) != 3 {
		t.Fatalf("splitMulti returned %d chunks, want 3", len(chunks))
	}
	rejoined := bytes.Join(chunks, nil)
	if string(rejoined) != "aaabbbcc" {
		t.Errorf("splitMulti rejoined = %q", rejoined)
	}
	if len(chunks[2]) != 2 {
		t.Errorf("last chunk length = %d, want 2", len(chunks[2]))
	}
}

func TestSplitClientHelloReassembles(t *testing.T) {
	data := bytes.Repeat([]byte{0xAB}, 128)
	for _, strategy := range []string{"half", "multi", "sni_split", "unknown"} {
		chunks := splitClientHello(data, strategy)
		if len(chunks) < 2 && strategy != "sni_split" {
			t.Errorf("strategy %q produced %d chunks, want >=2", strategy, len(chunks))
		}
		if !bytes.Equal(bytes.Join(chunks, nil), data) {
			t.Errorf("strategy %q did not preserve bytes", strategy)
		}
	}
}

// buildClientHelloWithSNI constructs a minimal but structurally valid TLS
// ClientHello record containing an SNI extension for the given hostname.
func buildClientHelloWithSNI(host string) []byte {
	var body []byte
	body = append(body, 0x03, 0x03)                 // client version TLS 1.2
	body = append(body, make([]byte, 32)...)        // random
	body = append(body, 0x00)                        // session id len = 0
	body = append(body, 0x00, 0x02, 0x00, 0x2f)      // cipher suites (len 2)
	body = append(body, 0x01, 0x00)                  // compression methods (len 1, null)

	// SNI extension
	hb := []byte(host)
	var sni []byte
	sni = append(sni, 0x00, byte(len(hb)+3)) // server_name_list length
	sni = append(sni, 0x00)                  // name_type host_name
	sni = append(sni, byte(len(hb)>>8), byte(len(hb)))
	sni = append(sni, hb...)

	var ext []byte
	ext = append(ext, 0x00, 0x00) // extension type SNI
	ext = append(ext, byte(len(sni)>>8), byte(len(sni)))
	ext = append(ext, sni...)

	body = append(body, byte(len(ext)>>8), byte(len(ext))) // extensions length
	body = append(body, ext...)

	// Handshake header: type(1)=ClientHello + length(3)
	hs := []byte{0x01, byte(len(body) >> 16), byte(len(body) >> 8), byte(len(body))}
	hs = append(hs, body...)

	// TLS record header: content type 0x16 handshake, version, length
	rec := []byte{0x16, 0x03, 0x01, byte(len(hs) >> 8), byte(len(hs))}
	rec = append(rec, hs...)
	return rec
}

func TestFindSNIHostnameOffset(t *testing.T) {
	host := "example.com"
	ch := buildClientHelloWithSNI(host)
	off := findSNIHostnameOffset(ch)
	if off < 0 {
		t.Fatalf("findSNIHostnameOffset returned -1")
	}
	if string(ch[off:off+len(host)]) != host {
		t.Errorf("hostname at offset = %q, want %q", string(ch[off:off+len(host)]), host)
	}
}

func TestFindSNIHostnameOffsetTooShort(t *testing.T) {
	if findSNIHostnameOffset([]byte{0x16, 0x03, 0x01}) != -1 {
		t.Errorf("expected -1 for short input")
	}
}

func TestSplitClientHelloSNIStrategy(t *testing.T) {
	ch := buildClientHelloWithSNI("example.com")
	chunks := splitClientHello(ch, "sni_split")
	if len(chunks) != 2 {
		t.Fatalf("sni_split produced %d chunks, want 2", len(chunks))
	}
	if !bytes.Equal(bytes.Join(chunks, nil), ch) {
		t.Errorf("sni_split did not preserve bytes")
	}
}

func TestFragmentWriteReassembles(t *testing.T) {
	data := buildClientHelloWithSNI("example.com")
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		_, _ = fragmentWrite(client, data, "multi", 0)
		client.Close()
	}()

	got, err := io.ReadAll(server)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("fragmentWrite delivered %d bytes, want %d (content mismatch)", len(got), len(data))
	}
}

func TestFragmentConnFragmentsClientHelloOnce(t *testing.T) {
	ch := buildClientHelloWithSNI("example.com")
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	fc := &fragmentConn{Conn: client, strategy: "multi"}
	go func() {
		fc.Write(ch)
		// A subsequent non-ClientHello write should pass through untouched.
		fc.Write([]byte("plain-data"))
		client.Close()
	}()

	got, err := io.ReadAll(server)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !fc.fragmented {
		t.Errorf("fragmentConn did not mark ClientHello as fragmented")
	}
	want := append(append([]byte{}, ch...), []byte("plain-data")...)
	if !bytes.Equal(got, want) {
		t.Errorf("reassembled stream mismatch")
	}
}
