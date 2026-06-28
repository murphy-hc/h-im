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
