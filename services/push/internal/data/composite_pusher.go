package data

import (
	"context"
	"os"

	pushpb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
	"github.com/murphy-hc/h-im/services/push/internal/biz"
)

// compositePusher routes push messages: FCM for Android, APNs for iOS.
type compositePusher struct {
	fcm  *fcmPusher
	apns *apnsPusher
}

// NewPusher creates a composite pusher. FCM and APNs are initialized from
// environment variables. If neither is configured, a stub is returned.
func NewPusher() biz.Pusher {
	fcm := newFCMPusher()
	apns := NewAPNSPusher()
	if fcm == nil && apns == nil {
		return &stubPusher{}
	}
	return &compositePusher{fcm: fcm, apns: apns}
}

func newFCMPusher() *fcmPusher {
	credFile := os.Getenv("FCM_CREDENTIALS")
	if credFile == "" {
		return nil
	}
	p, err := NewFCMPusher(credFile)
	if err != nil {
		return nil
	}
	return p
}

func (p *compositePusher) Send(ctx context.Context, tokens []string, platform int32, title, body string, data map[string]string) error {
	switch platform {
	case int32(pushpb.DevicePlatform_DEVICE_PLATFORM_IOS):
		if p.apns != nil {
			return p.apns.Send(ctx, tokens, platform, title, body, data)
		}
		// Fall back to FCM for iOS if APNs not configured
		if p.fcm != nil {
			return p.fcm.Send(ctx, tokens, platform, title, body, data)
		}
	default:
		if p.fcm != nil {
			return p.fcm.Send(ctx, tokens, platform, title, body, data)
		}
	}
	return nil
}

func (p *compositePusher) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	if p.fcm != nil {
		return p.fcm.SendToTopic(ctx, topic, title, body, data)
	}
	return nil
}
