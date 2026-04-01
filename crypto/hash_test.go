package crypto_test

import (
	"testing"

	"github.com/weprodev/go-pkg/crypto"
)

func TestHashAndCheckSecret(t *testing.T) {
	secret := "super_secret_password_123!"

	hash, err := crypto.HashSecret(secret)
	if err != nil {
		t.Fatalf("failed to hash secret: %v", err)
	}

	if crypto.CheckSecretHash("wrong_password", hash) {
		t.Error("expected wrong password to fail check")
	}

	if !crypto.CheckSecretHash(secret, hash) {
		t.Error("expected correct secret to pass check")
	}
}
