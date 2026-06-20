package service

import (
	"context"

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
	conns := s.cm.GetConns(req.UserId)
	if len(conns) == 0 {
		return &gatewayv1.SendToUserResponse{Success: false}, nil
	}
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	for _, conn := range conns {
		conn.WriteMessage(websocket.BinaryMessage, msg)
	}
	return &gatewayv1.SendToUserResponse{Success: true}, nil
}

func (s *GatewayGrpcService) BroadcastToGroup(ctx context.Context, req *gatewayv1.BroadcastToGroupRequest) (*gatewayv1.BroadcastToGroupResponse, error) {
	exclude := make(map[string]bool, len(req.ExcludeUserIds))
	for _, uid := range req.ExcludeUserIds { exclude[uid] = true }
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	var delivered int32
	for _, uid := range s.cm.GetGroupMembers(req.GroupId) {
		if exclude[uid] { continue }
		for _, conn := range s.cm.GetConns(uid) {
			conn.WriteMessage(websocket.BinaryMessage, msg)
			delivered++
		}
	}
	return &gatewayv1.BroadcastToGroupResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) BroadcastToChatroom(ctx context.Context, req *gatewayv1.BroadcastToChatroomRequest) (*gatewayv1.BroadcastToChatroomResponse, error) {
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	var delivered int32
	for _, uid := range s.cm.GetRoomMembers(req.RoomId) {
		for _, conn := range s.cm.GetConns(uid) {
			conn.WriteMessage(websocket.BinaryMessage, msg)
			delivered++
		}
	}
	return &gatewayv1.BroadcastToChatroomResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) SendCommand(ctx context.Context, req *gatewayv1.SendCommandRequest) (*gatewayv1.SendCommandResponse, error) {
	conns := s.cm.KickUser(req.UserId)
	if len(conns) == 0 {
		return &gatewayv1.SendCommandResponse{Success: false}, nil
	}
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, req.Command)
	for _, conn := range conns {
		conn.WriteMessage(websocket.CloseMessage, closeMsg)
		conn.Close()
	}
	return &gatewayv1.SendCommandResponse{Success: true}, nil
}
