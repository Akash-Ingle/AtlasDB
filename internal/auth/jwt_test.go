package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTRoundTrip(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	email := "test@atlasdb.dev"
	role := "editor"

	pair, err := mgr.GenerateTokenPair(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("access token is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("refresh token is empty")
	}
	if pair.ExpiresIn != 900 { // 15 min = 900s
		t.Errorf("ExpiresIn = %d, want 900", pair.ExpiresIn)
	}

	// Validate access token
	claims, err := mgr.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email = %q, want %q", claims.Email, email)
	}
	if claims.Role != role {
		t.Errorf("Role = %q, want %q", claims.Role, role)
	}

	// Validate refresh token
	refreshUserID, err := mgr.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error: %v", err)
	}
	if refreshUserID != userID {
		t.Errorf("refresh token user ID = %v, want %v", refreshUserID, userID)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	mgr := NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	_, err := mgr.ValidateAccessToken("invalid.token.string")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	mgr1 := NewJWTManager("secret-1", 15*time.Minute, 7*24*time.Hour)
	mgr2 := NewJWTManager("secret-2", 15*time.Minute, 7*24*time.Hour)

	pair, _ := mgr1.GenerateTokenPair(uuid.New(), "test@test.com", "viewer")

	_, err := mgr2.ValidateAccessToken(pair.AccessToken)
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}
