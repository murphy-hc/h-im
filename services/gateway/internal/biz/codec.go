package biz

import (
	"encoding/binary"
	"fmt"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"google.golang.org/protobuf/proto"
)

const HeaderSize = 9

const CurrentVersion uint8 = 1

// Encode serializes a WS frame: 1 byte version + 4 bytes frameType (big-endian) + 4 bytes payloadLen (big-endian) + payload.
func Encode(version uint8, frameType gatewayv1.FrameType, msg proto.Message) ([]byte, error) {
	var payload []byte
	if msg != nil {
		var err error
		payload, err = proto.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
	}
	buf := make([]byte, HeaderSize+len(payload))
	buf[0] = version
	binary.BigEndian.PutUint32(buf[1:5], uint32(frameType))
	binary.BigEndian.PutUint32(buf[5:9], uint32(len(payload)))
	copy(buf[9:], payload)
	return buf, nil
}

// Decode parses a WS frame and returns the version, frameType, payload bytes, and any error.
func Decode(frame []byte) (version uint8, frameType gatewayv1.FrameType, payload []byte, err error) {
	if len(frame) < HeaderSize {
		return 0, 0, nil, fmt.Errorf("frame too short: %d < %d", len(frame), HeaderSize)
	}
	version = frame[0]
	if version != CurrentVersion {
		return version, 0, nil, fmt.Errorf("unsupported version: %d", version)
	}
	frameType = gatewayv1.FrameType(binary.BigEndian.Uint32(frame[1:5]))
	payloadLen := binary.BigEndian.Uint32(frame[5:9])
	if uint32(len(frame)) < HeaderSize+payloadLen {
		return 0, 0, nil, fmt.Errorf("frame length mismatch")
	}
	payload = frame[9 : 9+payloadLen]
	return version, frameType, payload, nil
}
