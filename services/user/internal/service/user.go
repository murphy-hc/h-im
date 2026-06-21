package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/services/user/internal/biz"
)

// UserService implements the UserService gRPC server.
type UserService struct {
	pb.UnimplementedUserServiceServer
	uc *biz.UserUseCase
}

// NewUserService creates a UserService.
func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{uc: uc}
}

// ReportHeartbeat records a successful heartbeat from a gateway.
func (s *UserService) ReportHeartbeat(ctx context.Context, req *pb.ReportHeartbeatRequest) (*pb.ReportHeartbeatResponse, error) {
	err := s.uc.ReportHeartbeat(ctx, req.UserId, req.DeviceId, req.GatewayAddr, req.Timestamp)
	if err != nil {
		return nil, err
	}
	return &pb.ReportHeartbeatResponse{}, nil
}

// ReportDisconnect records a device disconnection.
func (s *UserService) ReportDisconnect(ctx context.Context, req *pb.ReportDisconnectRequest) (*pb.ReportDisconnectResponse, error) {
	err := s.uc.ReportDisconnect(ctx, req.UserId, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return &pb.ReportDisconnectResponse{}, nil
}

// GetUserOnline returns all online devices for a user.
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
