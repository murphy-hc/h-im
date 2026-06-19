package service

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

type GatewayGrpcService struct {
	gatewayv1.UnimplementedGatewayServiceServer
	cm biz.ConnManager
}

func NewGatewayGrpcService(cm biz.ConnManager) *GatewayGrpcService {
	return &GatewayGrpcService{cm: cm}
}

func (s *GatewayGrpcService) SendToUser(ctx context.Context, req *gatewayv1.SendToUserRequest) (*gatewayv1.SendToUserResponse, error) {
	conn, ok := s.cm.GetConn(req.UserId)
	if !ok {
		return &gatewayv1.SendToUserResponse{Success: false}, nil
	}
	ft := gatewayv1.FrameType(req.FrameType)
	frame, err := biz.Encode(biz.CurrentVersion, ft, nil)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	msg := make([]byte, len(frame)+len(req.Payload))
	copy(msg, frame)
	copy(msg[len(frame):], req.Payload)
	if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		return &gatewayv1.SendToUserResponse{Success: false}, nil
	}
	return &gatewayv1.SendToUserResponse{Success: true}, nil
}

func (s *GatewayGrpcService) BroadcastToGroup(ctx context.Context, req *gatewayv1.BroadcastToGroupRequest) (*gatewayv1.BroadcastToGroupResponse, error) {
	members := s.cm.GetGroupMembers(req.GroupId)
	ft := gatewayv1.FrameType(req.FrameType)
	frame, _ := biz.Encode(biz.CurrentVersion, ft, nil)
	msg := make([]byte, len(frame)+len(req.Payload))
	copy(msg, frame)
	copy(msg[len(frame):], req.Payload)
	exclude := make(map[string]bool)
	for _, uid := range req.ExcludeUserIds {
		exclude[uid] = true
	}
	var delivered int32
	for _, uid := range members {
		if exclude[uid] {
			continue
		}
		conn, ok := s.cm.GetConn(uid)
		if ok {
			conn.WriteMessage(websocket.BinaryMessage, msg)
			delivered++
		}
	}
	return &gatewayv1.BroadcastToGroupResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) BroadcastToChatroom(ctx context.Context, req *gatewayv1.BroadcastToChatroomRequest) (*gatewayv1.BroadcastToChatroomResponse, error) {
	members := s.cm.GetRoomMembers(req.RoomId)
	ft := gatewayv1.FrameType(req.FrameType)
	frame, _ := biz.Encode(biz.CurrentVersion, ft, nil)
	msg := make([]byte, len(frame)+len(req.Payload))
	copy(msg, frame)
	copy(msg[len(frame):], req.Payload)
	var delivered int32
	for _, uid := range members {
		if conn, ok := s.cm.GetConn(uid); ok {
			conn.WriteMessage(websocket.BinaryMessage, msg)
			delivered++
		}
	}
	return &gatewayv1.BroadcastToChatroomResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) SendCommand(ctx context.Context, req *gatewayv1.SendCommandRequest) (*gatewayv1.SendCommandResponse, error) {
	conn, ok := s.cm.GetConn(req.UserId)
	if !ok {
		return &gatewayv1.SendCommandResponse{Success: false}, nil
	}
	frame, _ := biz.Encode(biz.CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR, nil)
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, req.Command))
	_ = frame
	_ = conn
	return &gatewayv1.SendCommandResponse{Success: true}, nil
}
