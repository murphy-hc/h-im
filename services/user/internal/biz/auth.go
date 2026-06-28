package biz

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidSignature   = errors.New("invalid token signature")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

// TokenClaims is the decoded token payload.
type TokenClaims struct {
	Signature string `json:"signature"`
	CurTime   int64  `json:"curTime"`
	TTL       int64  `json:"ttl"`
}

// AuthUseCase handles token validation.
type AuthUseCase struct {
	repo UserRepo
}

// NewAuthUseCase creates an AuthUseCase.
func NewAuthUseCase(repo UserRepo) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}

// ValidateAppToken validates a token signed by the business app.
func (uc *AuthUseCase) ValidateAppToken(ctx context.Context, appID, userID, token string) error {
	app, err := uc.repo.FindAppByID(ctx, appID)
	if err != nil {
		return err
	}
	return verifyAppToken(app.AppSecret, userID, token)
}

func verifyAppToken(appSecret, userID, token string) error {
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return ErrInvalidTokenFormat
	}
	var claims TokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return ErrInvalidTokenFormat
	}
	expected := sign(appSecret, userID, claims.CurTime, claims.TTL)
	if expected != claims.Signature {
		return ErrInvalidSignature
	}
	if claims.CurTime+claims.TTL < time.Now().UnixMilli() {
		return ErrTokenExpired
	}
	return nil
}

func sign(appSecret, userID string, curTime, ttl int64) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(fmt.Sprintf("%s%d%d", userID, curTime, ttl)))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
