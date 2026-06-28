package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
	"github.com/murphy-hc/h-im/services/push/internal/biz"
)

type PushService struct {
	pb.UnimplementedPushServiceServer
	uc *biz.PushUseCase
}

func NewPushService(uc *biz.PushUseCase) *PushService { return &PushService{uc: uc} }

func (s *PushService) RegisterDevice(ctx context.Context, req *pb.RegisterDeviceRequest) (*pb.RegisterDeviceResponse, error) {
	dev := req.GetDevice()
	err := s.uc.RegisterDevice(ctx, req.UserId, &biz.DeviceInfo{
		DeviceID: dev.GetDeviceId(), DeviceToken: dev.GetDeviceToken(), Platform: int32(dev.GetPlatform()),
	})
	if err != nil { return nil, err }
	return &pb.RegisterDeviceResponse{Success: true}, nil
}

func (s *PushService) UnregisterDevice(ctx context.Context, req *pb.UnregisterDeviceRequest) (*pb.UnregisterDeviceResponse, error) {
	err := s.uc.UnregisterDevice(ctx, req.DeviceId)
	if err != nil { return nil, err }
	return &pb.UnregisterDeviceResponse{Success: true}, nil
}

func (s *PushService) PushToUser(ctx context.Context, req *pb.PushToUserRequest) (*pb.PushToUserResponse, error) {
	err := s.uc.PushToUser(ctx, req.UserId, req.Title, req.Body, []byte(req.Payload))
	if err != nil { return nil, err }
	return &pb.PushToUserResponse{Success: true}, nil
}

func (s *PushService) PushToTopic(ctx context.Context, req *pb.PushToTopicRequest) (*pb.PushToTopicResponse, error) {
	err := s.uc.PushToTopic(ctx, req.Topic, req.Title, req.Body, []byte(req.Payload))
	if err != nil { return nil, err }
	return &pb.PushToTopicResponse{Success: true}, nil
}
