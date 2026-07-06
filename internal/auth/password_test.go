package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "securePassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if hash == password {
		t.Error("hash should not equal plaintext")
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword() returned false for correct password")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Error("CheckPassword() returned true for wrong password")
	}
}

func TestHashPasswordTooShort(t *testing.T) {
	_, err := HashPassword("short")
	if err == nil {
		t.Error("expected error for short password")
	}
}
