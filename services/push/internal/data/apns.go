package data

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/go-kratos/kratos/v2/log"
)

// apnsPusher sends push notifications directly to Apple Push Notification service.
type apnsPusher struct {
	client      *http.Client
	keyID       string
	teamID      string
	privateKey  *ecdsa.PrivateKey
	bundleID    string
	production  bool
}

// NewAPNSPusher creates an APNs pusher from environment variables:
//
//	APNS_KEY_ID       — Key ID from Apple Developer
//	APNS_TEAM_ID      — Team ID
//	APNS_KEY_PATH     — path to .p8 private key file
//	APNS_BUNDLE_ID    — app bundle ID
//	APNS_PRODUCTION   — "true" for production, defaults to sandbox
func NewAPNSPusher() *apnsPusher {
	keyID := os.Getenv("APNS_KEY_ID")
	teamID := os.Getenv("APNS_TEAM_ID")
	keyPath := os.Getenv("APNS_KEY_PATH")
	bundleID := os.Getenv("APNS_BUNDLE_ID")
	if keyID == "" || teamID == "" || keyPath == "" || bundleID == "" {
		log.Warnf("APNs not configured (missing env vars: APNS_KEY_ID, APNS_TEAM_ID, APNS_KEY_PATH, APNS_BUNDLE_ID)")
		return nil
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		log.Errorf("apns: read key: %v", err)
		return nil
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		log.Errorf("apns: failed to decode PEM")
		return nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("apns: parse key: %v", err)
		return nil
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		log.Errorf("apns: not an EC private key")
		return nil
	}
	return &apnsPusher{
		client:     &http.Client{Timeout: 10 * time.Second},
		keyID:      keyID,
		teamID:     teamID,
		privateKey: ecKey,
		bundleID:   bundleID,
		production: os.Getenv("APNS_PRODUCTION") == "true",
	}
}

func (p *apnsPusher) Send(ctx context.Context, tokens []string, platform int32, title, body string, data map[string]string) error {
	if p == nil {
		return nil
	}
	token, err := p.jwt()
	if err != nil {
		return err
	}
	aps := map[string]interface{}{
		"aps": map[string]interface{}{
			"alert": map[string]string{"title": title, "body": body},
			"sound": "default",
			"badge": 1,
		},
	}
	for k, v := range data {
		aps[k] = v
	}
	payload, _ := json.Marshal(aps)
	host := "api.sandbox.push.apple.com"
	if p.production {
		host = "api.push.apple.com"
	}
	for _, deviceToken := range tokens {
		req, _ := http.NewRequestWithContext(ctx, "POST",
			fmt.Sprintf("https://%s/3/device/%s", host, deviceToken),
			bytes.NewReader(payload))
		req.Header.Set("authorization", "bearer "+token)
		req.Header.Set("apns-topic", p.bundleID)
		req.Header.Set("apns-push-type", "alert")
		req.Header.Set("apns-expiration", "0")

		resp, err := p.client.Do(req)
		if err != nil {
			return fmt.Errorf("apns send: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("apns: %s: %s", resp.Status, string(body))
		}
	}
	return nil
}

func (p *apnsPusher) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	// APNs doesn't support topic messaging natively; delegate to per-device push.
	// This method exists to satisfy the Pusher interface.
	log.Context(ctx).Warnf( "apns: SendToTopic not supported, use per-device push")
	return nil
}

func (p *apnsPusher) jwt() (string, error) {
	claims := jwt.MapClaims{
		"iss": p.teamID,
		"iat": time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	t.Header["kid"] = p.keyID
	return t.SignedString(p.privateKey)
}
