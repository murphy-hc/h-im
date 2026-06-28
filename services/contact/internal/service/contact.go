package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/contact/v1"
	"github.com/murphy-hc/h-im/services/contact/internal/biz"
)

type ContactService struct {
	pb.UnimplementedContactServiceServer
	uc *biz.ContactUseCase
}

func NewContactService(uc *biz.ContactUseCase) *ContactService {
	return &ContactService{uc: uc}
}

func (s *ContactService) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error) {
	reqID, err := s.uc.SendFriendRequest(ctx, req.FromUser, req.ToUser, req.Message)
	if err != nil { return nil, err }
	return &pb.SendFriendRequestResponse{RequestId: reqID}, nil
}

func (s *ContactService) AcceptFriendRequest(ctx context.Context, req *pb.AcceptFriendRequestRequest) (*pb.AcceptFriendRequestResponse, error) {
	err := s.uc.AcceptFriendRequest(ctx, req.RequestId)
	if err != nil { return nil, err }
	return &pb.AcceptFriendRequestResponse{Success: true}, nil
}

func (s *ContactService) RejectFriendRequest(ctx context.Context, req *pb.RejectFriendRequestRequest) (*pb.RejectFriendRequestResponse, error) {
	err := s.uc.RejectFriendRequest(ctx, req.RequestId)
	if err != nil { return nil, err }
	return &pb.RejectFriendRequestResponse{Success: true}, nil
}

func (s *ContactService) RemoveFriend(ctx context.Context, req *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	err := s.uc.RemoveFriend(ctx, req.UserId, req.FriendId)
	if err != nil { return nil, err }
	return &pb.RemoveFriendResponse{Success: true}, nil
}

func (s *ContactService) BlockUser(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	err := s.uc.BlockUser(ctx, req.UserId, req.BlockId)
	if err != nil { return nil, err }
	return &pb.BlockUserResponse{Success: true}, nil
}

func (s *ContactService) UnblockUser(ctx context.Context, req *pb.UnblockUserRequest) (*pb.UnblockUserResponse, error) {
	err := s.uc.UnblockUser(ctx, req.UserId, req.BlockId)
	if err != nil { return nil, err }
	return &pb.UnblockUserResponse{Success: true}, nil
}

func (s *ContactService) GetFriends(ctx context.Context, req *pb.GetFriendsRequest) (*pb.GetFriendsResponse, error) {
	offset, limit := int32(0), int32(50)
	if pg := req.GetPagination(); pg != nil {
		offset, limit = int32(pg.GetPage()), int32(pg.GetPageSize())
	}
	friends, err := s.uc.GetFriends(ctx, req.UserId, offset, limit)
	if err != nil { return nil, err }
	pbFriends := make([]*pb.FriendInfo, len(friends))
	for i, f := range friends {
		pbFriends[i] = &pb.FriendInfo{UserId: f.UserID, Status: pb.FriendStatus(f.Status)}
	}
	return &pb.GetFriendsResponse{Friends: pbFriends}, nil
}

func (s *ContactService) GetFriendRequests(ctx context.Context, req *pb.GetFriendRequestsRequest) (*pb.GetFriendRequestsResponse, error) {
	requests, err := s.uc.GetFriendRequests(ctx, req.UserId)
	if err != nil { return nil, err }
	pbReqs := make([]*pb.FriendRequest, len(requests))
	for i, r := range requests {
		pbReqs[i] = &pb.FriendRequest{RequestId: r.RequestID, FromUser: r.FromUser, ToUser: r.ToUser, Message: r.Message}
	}
	return &pb.GetFriendRequestsResponse{Requests: pbReqs}, nil
}
