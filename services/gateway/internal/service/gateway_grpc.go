package service

import (
	"context"

	"github.com/coder/websocket"
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

func writeToConns(ctx context.Context, conns []*websocket.Conn, msg []byte) (delivered int32) {
	for _, c := range conns {
		if err := c.Write(ctx, websocket.MessageBinary, msg); err == nil {
			delivered++
		}
	}
	return
}

func (s *GatewayGrpcService) SendToUser(ctx context.Context, req *gatewayv1.SendToUserRequest) (*gatewayv1.SendToUserResponse, error) {
	conns, _ := s.cm.GetConns(req.UserId)
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	writeToConns(ctx, conns, msg)
	return &gatewayv1.SendToUserResponse{Success: len(conns) > 0}, nil
}

func (s *GatewayGrpcService) BroadcastToGroup(ctx context.Context, req *gatewayv1.BroadcastToGroupRequest) (*gatewayv1.BroadcastToGroupResponse, error) {
	exclude := make(map[string]bool, len(req.ExcludeUserIds))
	for _, uid := range req.ExcludeUserIds {
		exclude[uid] = true
	}
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	var delivered int32
	members, _ := s.cm.GetGroupMembers(req.GroupId)
	for _, uid := range members {
		if exclude[uid] {
			continue
		}
		conns, _ := s.cm.GetConns(uid)
		delivered += writeToConns(ctx, conns, msg)
	}
	return &gatewayv1.BroadcastToGroupResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) BroadcastToChatroom(ctx context.Context, req *gatewayv1.BroadcastToChatroomRequest) (*gatewayv1.BroadcastToChatroomResponse, error) {
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	var delivered int32
	members, _ := s.cm.GetRoomMembers(req.RoomId)
	for _, uid := range members {
		conns, _ := s.cm.GetConns(uid)
		delivered += writeToConns(ctx, conns, msg)
	}
	return &gatewayv1.BroadcastToChatroomResponse{DeliveredCount: delivered}, nil
}

func (s *GatewayGrpcService) SendCommand(ctx context.Context, req *gatewayv1.SendCommandRequest) (*gatewayv1.SendCommandResponse, error) {
	conns, _ := s.cm.KickUser(req.UserId)
	for _, c := range conns {
		c.Close(websocket.StatusNormalClosure, req.Command)
	}
	return &gatewayv1.SendCommandResponse{Success: len(conns) > 0}, nil
}
