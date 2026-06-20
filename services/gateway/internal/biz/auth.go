package biz

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidSignature  = errors.New("invalid token signature")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

// TokenClaims is the decoded token payload.
type TokenClaims struct {
	Signature string `json:"signature"`
	CurTime   int64  `json:"curTime"`
	TTL       int64  `json:"ttl"`
}

// VerifyAppToken validates the token signed by the business app.
func VerifyAppToken(appSecret, userID, token string) error {
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
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%s%s%d%d", appSecret, userID, curTime, ttl))))
}
