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

func writeToConns(conns []*websocket.Conn, msg []byte) int32 {
	var n int32
	for _, c := range conns {
		c.WriteMessage(websocket.BinaryMessage, msg)
		n++
	}
	return n
}

func (s *GatewayGrpcService) SendToUser(ctx context.Context, req *gatewayv1.SendToUserRequest) (*gatewayv1.SendToUserResponse, error) {
	conns, _ := s.cm.GetConns(req.UserId)
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	writeToConns(conns, msg)
	return &gatewayv1.SendToUserResponse{Success: len(conns) > 0}, nil
}

func (s *GatewayGrpcService) BroadcastToGroup(ctx context.Context, req *gatewayv1.BroadcastToGroupRequest) (*gatewayv1.BroadcastToGroupResponse, error) {
	exclude := make(map[string]bool, len(req.ExcludeUserIds))
	for _, uid := range req.ExcludeUserIds { exclude[uid] = true }
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	var delivered int32
	members, _ := s.cm.GetGroupMembers(req.GroupId)
	for _, uid := range members {
		if exclude[uid] { continue }
		conns, _ := s.cm.GetConns(uid)
		delivered += writeToConns(conns, msg)
	}
	return &gatewayv1.BroadcastToGroupResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) BroadcastToChatroom(ctx context.Context, req *gatewayv1.BroadcastToChatroomRequest) (*gatewayv1.BroadcastToChatroomResponse, error) {
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	var delivered int32
	members, _ := s.cm.GetRoomMembers(req.RoomId)
	for _, uid := range members {
		conns, _ := s.cm.GetConns(uid)
		delivered += writeToConns(conns, msg)
	}
	return &gatewayv1.BroadcastToChatroomResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) SendCommand(ctx context.Context, req *gatewayv1.SendCommandRequest) (*gatewayv1.SendCommandResponse, error) {
	conns, _ := s.cm.KickUser(req.UserId)
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, req.Command)
	writeToConns(conns, closeMsg)
	return &gatewayv1.SendCommandResponse{Success: len(conns) > 0}, nil
}
