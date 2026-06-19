package biz_test

import (
	"testing"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"google.golang.org/protobuf/proto"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := &gatewayv1.AuthRequest{Token: "test-token"}
	frame, err := biz.Encode(1, gatewayv1.FrameType_FRAME_TYPE_AUTH_REQ, original)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	version, ft, payload, err := biz.Decode(frame)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if version != 1 {
		t.Errorf("version = %d, want 1", version)
	}
	if ft != gatewayv1.FrameType_FRAME_TYPE_AUTH_REQ {
		t.Errorf("wrong frame type")
	}
	var decoded gatewayv1.AuthRequest
	if err := proto.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Token != original.Token {
		t.Errorf("token mismatch")
	}
}

func TestDecodeInvalidFrame(t *testing.T) {
	_, _, _, err := biz.Decode([]byte{})
	if err == nil {
		t.Fatal("expected error for empty frame")
	}
}

func TestEncodeEmptyPayload(t *testing.T) {
	frame, err := biz.Encode(1, gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if len(frame) != 9 {
		t.Errorf("expected 9 bytes (1+4+4+0), got %d", len(frame))
	}
	_, ft, payload, _ := biz.Decode(frame)
	if ft != gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT {
		t.Errorf("wrong frame type")
	}
	if len(payload) != 0 {
		t.Errorf("expected empty payload")
	}
}
