package crypto_test

import (
	"bytes"
	"testing"

	"github.com/thiemotorres/umgebung/internal/crypto"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("super secret value")

	ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	got, err := crypto.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("got %q, want %q", got, plaintext)
	}
}

func TestDeriveKey(t *testing.T) {
	salt := make([]byte, 16)
	key1 := crypto.DeriveKey("password", salt)
	key2 := crypto.DeriveKey("password", salt)
	key3 := crypto.DeriveKey("wrong", salt)

	if !bytes.Equal(key1, key2) {
		t.Fatal("same password+salt should produce same key")
	}
	if bytes.Equal(key1, key3) {
		t.Fatal("different password should produce different key")
	}
	if len(key1) != 32 {
		t.Fatalf("key should be 32 bytes, got %d", len(key1))
	}
}

func TestGenerateSalt(t *testing.T) {
	s1 := crypto.GenerateSalt()
	s2 := crypto.GenerateSalt()
	if bytes.Equal(s1, s2) {
		t.Fatal("salts should be random")
	}
	if len(s1) != 16 {
		t.Fatalf("salt should be 16 bytes, got %d", len(s1))
	}
}
