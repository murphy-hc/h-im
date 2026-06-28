package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/chatroom/v1"
	commonv1 "github.com/murphy-hc/h-im/gen/go/him/common/v1"
	"github.com/murphy-hc/h-im/services/chatroom/internal/biz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatroomService struct {
	pb.UnimplementedChatroomServiceServer
	uc *biz.ChatroomUseCase
}

func NewChatroomService(uc *biz.ChatroomUseCase) *ChatroomService {
	return &ChatroomService{uc: uc}
}

func (s *ChatroomService) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	room, err := s.uc.CreateRoom(ctx, req.Name, req.OwnerId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateRoomResponse{RoomId: room.RoomID}, nil
}

func (s *ChatroomService) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	err := s.uc.JoinRoom(ctx, req.RoomId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.JoinRoomResponse{}, nil
}

func (s *ChatroomService) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*pb.LeaveRoomResponse, error) {
	err := s.uc.LeaveRoom(ctx, req.RoomId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.LeaveRoomResponse{}, nil
}

func (s *ChatroomService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "use WebSocket gateway to send chatroom messages")
}

func (s *ChatroomService) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	page := int32(1)
	pageSize := int32(20)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
	}
	offset := (page - 1) * pageSize
	msgs, total, err := s.uc.GetMessages(ctx, req.RoomId, offset, pageSize)
	if err != nil {
		return nil, err
	}
	pbMsgs := make([]*pb.ChatroomMessage, 0, len(msgs))
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, &pb.ChatroomMessage{
			MessageId: m.ServerID,
			RoomId:    m.RoomID,
			SenderId:  m.SenderID,
			CreatedAt: m.CreateTime,
		})
	}
	totalPage := int32(0)
	if pageSize > 0 {
		totalPage = int32(total) / pageSize
		if int32(total)%pageSize > 0 {
			totalPage++
		}
	}
	return &pb.GetMessagesResponse{
		Messages: pbMsgs,
		Pagination: &commonv1.PaginationResponse{
			Page: page, PageSize: pageSize, Total: total, TotalPage: totalPage,
		},
	}, nil
}
