package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/services/user/internal/biz"
)

// UserService implements the UserService gRPC server.
type UserService struct {
	pb.UnimplementedUserServiceServer
	uc     *biz.UserUseCase
	authUC *biz.AuthUseCase
}

// NewUserService creates a UserService.
func NewUserService(uc *biz.UserUseCase, authUC *biz.AuthUseCase) *UserService {
	return &UserService{uc: uc, authUC: authUC}
}

func (s *UserService) ReportHeartbeat(ctx context.Context, req *pb.ReportHeartbeatRequest) (*pb.ReportHeartbeatResponse, error) {
	err := s.uc.ReportHeartbeat(ctx, req.UserId, req.DeviceId, req.GatewayAddr, req.Timestamp)
	if err != nil {
		return nil, err
	}
	return &pb.ReportHeartbeatResponse{}, nil
}

func (s *UserService) ReportDisconnect(ctx context.Context, req *pb.ReportDisconnectRequest) (*pb.ReportDisconnectResponse, error) {
	err := s.uc.ReportDisconnect(ctx, req.UserId, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return &pb.ReportDisconnectResponse{}, nil
}

func (s *UserService) GetUserOnline(ctx context.Context, req *pb.GetUserOnlineRequest) (*pb.GetUserOnlineResponse, error) {
	devices, err := s.uc.GetUserOnline(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	pbDevices := make([]*pb.DeviceOnlineInfo, 0, len(devices))
	for _, d := range devices {
		pbDevices = append(pbDevices, &pb.DeviceOnlineInfo{
			DeviceId:      d.DeviceID,
			GatewayAddr:   d.GatewayAddr,
			LastHeartbeat: d.LastHeartbeat,
		})
	}
	return &pb.GetUserOnlineResponse{Devices: pbDevices}, nil
}

func (s *UserService) ValidateAppToken(ctx context.Context, req *pb.ValidateAppTokenRequest) (*pb.ValidateAppTokenResponse, error) {
	err := s.authUC.ValidateAppToken(ctx, req.AppId, req.UserId, req.Token)
	if err != nil {
		return &pb.ValidateAppTokenResponse{Valid: false}, nil
	}
	return &pb.ValidateAppTokenResponse{Valid: true}, nil
}

func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, err := s.uc.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{UserId: userID}, nil
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	access, refresh, expiresAt, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	u, err := s.uc.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.GetProfileResponse{User: &pb.User{
		UserId: u.UserID, Nickname: u.Nickname, Avatar: u.Avatar,
	}}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	err := s.uc.UpdateProfile(ctx, req.UserId, req.Nickname, req.Avatar)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateProfileResponse{}, nil
}

func (s *UserService) BatchGetUsers(ctx context.Context, req *pb.BatchGetUsersRequest) (*pb.BatchGetUsersResponse, error) {
	users, err := s.uc.BatchGetUsers(ctx, req.UserIds)
	if err != nil {
		return nil, err
	}
	pbUsers := make([]*pb.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, &pb.User{
			UserId: u.UserID, Nickname: u.Nickname, Avatar: u.Avatar,
		})
	}
	return &pb.BatchGetUsersResponse{Users: pbUsers}, nil
}
