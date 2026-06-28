package data

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/api/option"
)

// fcmPusher sends push via Firebase Cloud Messaging.
type fcmPusher struct {
	app *firebase.App
}

// NewFCMPusher creates an FCM-based pusher from a service account JSON file.
func NewFCMPusher(credentialsFile string) (*fcmPusher, error) {
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}
	return &fcmPusher{app: app}, nil
}

func (p *fcmPusher) Send(ctx context.Context, tokens []string, platform int32, title, body string, data map[string]string) error {
	client, err := p.app.Messaging(ctx)
	if err != nil {
		return err
	}
	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data: data,
		Android: &messaging.AndroidConfig{Priority: "high"},
		APNS:    &messaging.APNSConfig{Payload: &messaging.APNSPayload{Aps: &messaging.Aps{Sound: "default"}}},
	}
	_, err = client.SendEachForMulticast(ctx, msg)
	return err
}

func (p *fcmPusher) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	client, err := p.app.Messaging(ctx)
	if err != nil {
		return err
	}
	msg := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data: data,
		Android: &messaging.AndroidConfig{Priority: "high"},
		APNS:    &messaging.APNSConfig{Payload: &messaging.APNSPayload{Aps: &messaging.Aps{Sound: "default"}}},
	}
	_, err = client.Send(ctx, msg)
	return err
}

// stubPusher logs push messages without sending (dev mode).
type stubPusher struct{}

func (p *stubPusher) Send(ctx context.Context, tokens []string, platform int32, title, body string, data map[string]string) error {
	log.Infof("[push-stub] send %d tokens: title=%q body=%q platform=%d", len(tokens), title, body, platform)
	return nil
}

func (p *stubPusher) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	log.Infof("[push-stub] send to topic %q: title=%q body=%q", topic, title, body)
	return nil
}
