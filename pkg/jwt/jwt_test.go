package jwt

import (
	"testing"
	"time"
)

func TestIssueAndValidate(t *testing.T) {
	mgr := NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	access, err := mgr.IssueAccessToken("user-123")
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	if access == "" {
		t.Fatal("empty access token")
	}

	userID, err := mgr.Validate(access)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("expected user-123, got %s", userID)
	}
}

func TestRefreshToken(t *testing.T) {
	mgr := NewManager("test-secret", 15*time.Minute, 24*time.Hour)

	refresh, err := mgr.IssueRefreshToken("user-456")
	if err != nil {
		t.Fatalf("IssueRefreshToken: %v", err)
	}

	userID, err := mgr.Validate(refresh)
	if err != nil {
		t.Fatalf("Validate refresh: %v", err)
	}
	if userID != "user-456" {
		t.Fatalf("expected user-456, got %s", userID)
	}
}

func TestInvalidToken(t *testing.T) {
	mgr := NewManager("secret-a", 1*time.Hour, 24*time.Hour)
	mgr2 := NewManager("secret-b", 1*time.Hour, 24*time.Hour)

	token, _ := mgr.IssueAccessToken("user-1")
	_, err := mgr2.Validate(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestExpiredToken(t *testing.T) {
	mgr := NewManager("secret", 1*time.Millisecond, 24*time.Hour)

	token, err := mgr.IssueAccessToken("user-1")
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	_, err = mgr.Validate(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestAccessTTL(t *testing.T) {
	mgr := NewManager("secret", 5*time.Minute, 30*time.Hour)
	if mgr.AccessTTL() != 5*time.Minute {
		t.Fatalf("expected 5m, got %v", mgr.AccessTTL())
	}
}
