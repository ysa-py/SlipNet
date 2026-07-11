package main

import (
	"bufio"
	"encoding/base64"
	"io"
	"net"
	"testing"
)

func TestSendPayloadPlaceholders(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	payload := "GET / HTTP/1.1[crlf]Host: [host]:[port][crlf][crlf]"
	go func() {
		_ = sendPayload(client, payload, "example.com", 8080)
		client.Close()
	}()

	got, err := io.ReadAll(server)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := "GET / HTTP/1.1\r\nHost: example.com:8080\r\n\r\n"
	if string(got) != want {
		t.Errorf("sendPayload resolved = %q, want %q", string(got), want)
	}
}

func TestSendPayloadCRLFVariants(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		_ = sendPayload(client, "a[cr]b[lf]c", "h", 1)
		client.Close()
	}()
	got, _ := io.ReadAll(server)
	if string(got) != "a\rb\nc" {
		t.Errorf("sendPayload [cr]/[lf] = %q, want %q", string(got), "a\rb\nc")
	}
}

func TestGenerateWSKey(t *testing.T) {
	k := generateWSKey()
	dec, err := base64.StdEncoding.DecodeString(k)
	if err != nil {
		t.Fatalf("not valid base64: %v", err)
	}
	if len(dec) != 16 {
		t.Errorf("decoded length = %d, want 16", len(dec))
	}
}

func TestWSConnRoundTrip(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	writer := newWSConn(a, nil)
	reader := newWSConn(b, bufio.NewReader(b))

	payloads := [][]byte{
		[]byte("short"),
		make([]byte, 300),   // 126 extended length path
		make([]byte, 70000), // 127 extended length path
	}
	for i := range payloads[2] {
		payloads[2][i] = byte(i % 251)
	}

	go func() {
		for _, p := range payloads {
			_, _ = writer.Write(p)
		}
	}()

	for i, want := range payloads {
		got := make([]byte, len(want))
		if _, err := io.ReadFull(reader, got); err != nil {
			t.Fatalf("read payload %d: %v", i, err)
		}
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("payload %d byte %d = %d, want %d", i, j, got[j], want[j])
			}
		}
	}
}

func TestWSConnWriteEmpty(t *testing.T) {
	a, _ := net.Pipe()
	defer a.Close()
	c := newWSConn(a, nil)
	n, err := c.Write(nil)
	if n != 0 || err != nil {
		t.Errorf("Write(nil) = %d,%v, want 0,nil", n, err)
	}
}

func TestBufferedConnRead(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	br := bufio.NewReader(b)
	bc := &bufferedConn{Conn: b, reader: br}

	go func() {
		a.Write([]byte("hello"))
		a.Close()
	}()

	got := make([]byte, 5)
	if _, err := io.ReadFull(bc, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("bufferedConn.Read = %q, want hello", string(got))
	}
}
