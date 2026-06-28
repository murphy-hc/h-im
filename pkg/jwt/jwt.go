// Package jwt provides helpers for issuing and validating JWTs.
package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims carries the standard JWT claims plus a UserID field.
type Claims struct {
	jwtlib.RegisteredClaims
	UserID string `json:"user_id"`
}

// Manager wraps a JWT secret and TTL configuration.
type Manager struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewManager creates a Manager.
func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// AccessTTL returns the configured access token TTL.
func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

// IssueAccessToken creates a short-lived access token for the given user.
func (m *Manager) IssueAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.accessTTL)),
		},
		UserID: userID,
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// IssueRefreshToken creates a long-lived refresh token.
func (m *Manager) IssueRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.refreshTTL)),
		},
		UserID: userID,
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Validate parses and validates a token string, returning the user ID if valid.
func (m *Manager) Validate(tokenStr string) (string, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwtlib.Token) (any, error) {
			return m.secret, nil
		},
	)
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", jwtlib.ErrSignatureInvalid
	}
	return claims.UserID, nil
}
