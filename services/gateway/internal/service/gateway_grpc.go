package service

import (
	"context"

	"github.com/coder/websocket"
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/rs/xid"
)

type GatewayGrpcService struct {
	gatewayv1.UnimplementedGatewayServiceServer
	cm          biz.ConnManager
	broadcaster biz.Broadcaster
}

func NewGatewayGrpcService(cm biz.ConnManager, broadcaster biz.Broadcaster) *GatewayGrpcService {
	return &GatewayGrpcService{cm: cm, broadcaster: broadcaster}
}

// HandleBroadcast delivers a Pub/Sub broadcast to local connections.
func (s *GatewayGrpcService) HandleBroadcast(ctx context.Context, bm *biz.BroadcastMsg) {
	s.deliverBroadcast(ctx, bm)
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
	conns, _ := s.cm.GetConns(ctx, req.UserId)
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(req.FrameType), req.Payload)
	delivered := writeToConns(ctx, conns, msg)
	return &gatewayv1.SendToUserResponse{Success: delivered > 0}, nil
}

func (s *GatewayGrpcService) BroadcastToGroup(ctx context.Context, req *gatewayv1.BroadcastToGroupRequest) (*gatewayv1.BroadcastToGroupResponse, error) {
	s.broadcast(ctx, biz.BroadcastTypeGroup, req.GroupId, req.FrameType, req.Payload)
	return &gatewayv1.BroadcastToGroupResponse{}, nil
}

func (s *GatewayGrpcService) BroadcastToChatroom(ctx context.Context, req *gatewayv1.BroadcastToChatroomRequest) (*gatewayv1.BroadcastToChatroomResponse, error) {
	s.broadcast(ctx, biz.BroadcastTypeRoom, req.RoomId, req.FrameType, req.Payload)
	return &gatewayv1.BroadcastToChatroomResponse{}, nil
}

func (s *GatewayGrpcService) broadcast(ctx context.Context, msgType int32, targetID string, frameType int32, payload []byte) {
	bm := &biz.BroadcastMsg{
		Type: msgType, TargetID: targetID, FrameType: frameType,
		Payload: payload, MsgID: xid.New().String(),
	}
	s.broadcaster.Publish(ctx, bm)
	s.deliverBroadcast(ctx, bm)
}

func (s *GatewayGrpcService) deliverBroadcast(ctx context.Context, bm *biz.BroadcastMsg) {
	msg := biz.BuildFrame(biz.CurrentVersion, uint32(bm.FrameType), bm.Payload)
	var memberIDs []string
	switch bm.Type {
	case biz.BroadcastTypeGroup:
		memberIDs, _ = s.cm.GetGroupMembers(ctx, bm.TargetID)
	case biz.BroadcastTypeRoom:
		memberIDs, _ = s.cm.GetRoomMembers(ctx, bm.TargetID)
	}
	for _, uid := range memberIDs {
		conns, _ := s.cm.GetConns(ctx, uid)
		writeToConns(ctx, conns, msg)
	}
}

func (s *GatewayGrpcService) JoinChatroom(ctx context.Context, req *gatewayv1.JoinChatroomRequest) (*gatewayv1.JoinChatroomResponse, error) {
	err := s.cm.JoinRoom(ctx, req.RoomId, req.UserId)
	if err != nil {
		return &gatewayv1.JoinChatroomResponse{Success: false}, nil
	}
	return &gatewayv1.JoinChatroomResponse{Success: true}, nil
}

func (s *GatewayGrpcService) LeaveChatroom(ctx context.Context, req *gatewayv1.LeaveChatroomRequest) (*gatewayv1.LeaveChatroomResponse, error) {
	err := s.cm.LeaveRoom(ctx, req.RoomId, req.UserId)
	if err != nil {
		return &gatewayv1.LeaveChatroomResponse{Success: false}, nil
	}
	return &gatewayv1.LeaveChatroomResponse{Success: true}, nil
}

func (s *GatewayGrpcService) SendCommand(ctx context.Context, req *gatewayv1.SendCommandRequest) (*gatewayv1.SendCommandResponse, error) {
	conns, _ := s.cm.KickUser(ctx, req.UserId)
	for _, c := range conns {
		c.Close(websocket.StatusNormalClosure, req.Command)
	}
	return &gatewayv1.SendCommandResponse{Success: len(conns) > 0}, nil
}

func (s *GatewayGrpcService) JoinGroup(ctx context.Context, req *gatewayv1.JoinGroupRequest) (*gatewayv1.JoinGroupResponse, error) {
	if err := s.cm.JoinGroup(ctx, req.GroupId, req.UserId); err != nil {
		return &gatewayv1.JoinGroupResponse{Success: false}, nil
	}
	return &gatewayv1.JoinGroupResponse{Success: true}, nil
}

func (s *GatewayGrpcService) LeaveGroup(ctx context.Context, req *gatewayv1.LeaveGroupRequest) (*gatewayv1.LeaveGroupResponse, error) {
	if err := s.cm.LeaveGroup(ctx, req.GroupId, req.UserId); err != nil {
		return &gatewayv1.LeaveGroupResponse{Success: false}, nil
	}
	return &gatewayv1.LeaveGroupResponse{Success: true}, nil
}
