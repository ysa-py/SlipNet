package main

import "testing"

// In the default build (no has_config_key tag, no -X main.configKey), the
// config key is unset, so decryptConfig must fail cleanly rather than panic.
func TestDecryptConfigNoKey(t *testing.T) {
	if configKeyBytes() != nil {
		t.Skip("build embeds a config key; skipping no-key path")
	}
	blob := append([]byte{encFormatVersion}, make([]byte, gcmIVLength+gcmTagLength)...)
	if _, err := decryptConfig(blob); err == nil {
		t.Errorf("expected error when no config key is set")
	}
}

func TestConfigKeyBytesInvalid(t *testing.T) {
	orig := configKey
	defer func() { configKey = orig }()

	configKey = ""
	if configKeyBytes() != nil {
		t.Errorf("empty configKey should yield nil")
	}

	configKey = "nothex!!"
	if configKeyBytes() != nil {
		t.Errorf("non-hex configKey should yield nil")
	}

	configKey = "abcd" // valid hex but wrong length (not 32 bytes)
	if configKeyBytes() != nil {
		t.Errorf("short configKey should yield nil")
	}
}

func TestDecryptConfigWithKeyErrors(t *testing.T) {
	orig := configKey
	defer func() { configKey = orig }()
	// 32-byte key expressed as hex (64 hex chars).
	configKey = "0000000000000000000000000000000000000000000000000000000000000000"
	if configKeyBytes() == nil {
		t.Fatalf("expected valid key bytes")
	}

	t.Run("empty data", func(t *testing.T) {
		if _, err := decryptConfig(nil); err == nil {
			t.Errorf("expected error for empty data")
		}
	})

	t.Run("bad version", func(t *testing.T) {
		if _, err := decryptConfig([]byte{0xFF, 1, 2, 3}); err == nil {
			t.Errorf("expected error for unsupported version byte")
		}
	})

	t.Run("too short", func(t *testing.T) {
		if _, err := decryptConfig([]byte{encFormatVersion, 0x00}); err == nil {
			t.Errorf("expected error for truncated data")
		}
	})

	t.Run("bad ciphertext", func(t *testing.T) {
		blob := append([]byte{encFormatVersion}, make([]byte, gcmIVLength+gcmTagLength)...)
		if _, err := decryptConfig(blob); err == nil {
			t.Errorf("expected decryption failure for garbage ciphertext")
		}
	})
}
