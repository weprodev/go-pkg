package crypto_test

import (
	"bytes"
	"testing"

	"github.com/weprodev/go-pkg/crypto"
)

func TestAESService_Validation(t *testing.T) {
	shortKey := []byte("short_key")
	if _, err := crypto.NewAESService(shortKey); err == nil {
		t.Fatal("expected error for non-32-byte key")
	}

	validKey := make([]byte, 32)
	if _, err := crypto.NewAESService(validKey); err != nil {
		t.Fatalf("unexpected error for 32-byte key: %v", err)
	}
}

func TestAESService_EncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	svc, err := crypto.NewAESService(key)
	if err != nil {
		t.Fatalf("failed to create AES service: %v", err)
	}

	plaintext := []byte("hello world this is a secret")
	ciphertext, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("ciphertext should not match plaintext")
	}

	decrypted, err := svc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestAESService_DecryptShortCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	svc, _ := crypto.NewAESService(key)

	// Minimal GCM nonce size is 12. Providing less than that should error out safely.
	shortData := []byte("too_short")
	_, err := svc.Decrypt(shortData)
	if err == nil {
		t.Fatal("expected error decoding short ciphertext")
	}
}
