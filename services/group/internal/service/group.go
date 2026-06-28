package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/group/v1"
	"github.com/murphy-hc/h-im/services/group/internal/biz"
)

type GroupService struct {
	pb.UnimplementedGroupServiceServer
	uc *biz.GroupUseCase
}

func NewGroupService(uc *biz.GroupUseCase) *GroupService {
	return &GroupService{uc: uc}
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	g, err := s.uc.CreateGroup(ctx, req.Name, req.OwnerId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateGroupResponse{GroupId: g.GroupID}, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, req *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	err := s.uc.JoinGroup(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.JoinGroupResponse{Success: true}, nil
}

func (s *GroupService) LeaveGroup(ctx context.Context, req *pb.LeaveGroupRequest) (*pb.LeaveGroupResponse, error) {
	err := s.uc.LeaveGroup(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.LeaveGroupResponse{Success: true}, nil
}

func (s *GroupService) DismissGroup(ctx context.Context, req *pb.DismissGroupRequest) (*pb.DismissGroupResponse, error) {
	err := s.uc.DismissGroup(ctx, req.GroupId, req.OwnerId)
	if err != nil {
		return nil, err
	}
	return &pb.DismissGroupResponse{Success: true}, nil
}

func (s *GroupService) GetGroupInfo(ctx context.Context, req *pb.GetGroupInfoRequest) (*pb.GetGroupInfoResponse, error) {
	g, err := s.uc.GetGroupInfo(ctx, req.GroupId)
	if err != nil {
		return nil, err
	}
	return &pb.GetGroupInfoResponse{Group: &pb.GroupInfo{
		GroupId: g.GroupID, Name: g.Name, OwnerId: g.OwnerID,
		Announcement: g.Announcement, MemberCount: g.MemberCount,
	}}, nil
}

func (s *GroupService) GetGroupMembers(ctx context.Context, req *pb.GetGroupMembersRequest) (*pb.GetGroupMembersResponse, error) {
	page, pageSize := int32(1), int32(50)
	if pg := req.GetPagination(); pg != nil {
		page, pageSize = int32(pg.GetPage()), int32(pg.GetPageSize())
	}
	members, err := s.uc.GetGroupMembers(ctx, req.GroupId, page, pageSize)
	if err != nil {
		return nil, err
	}
	pbMembers := make([]*pb.GroupMember, len(members))
	for i, m := range members {
		pbMembers[i] = &pb.GroupMember{UserId: m.UserID, Role: pb.GroupRole(m.Role)}
	}
	return &pb.GetGroupMembersResponse{Members: pbMembers}, nil
}

func (s *GroupService) SetMemberRole(ctx context.Context, req *pb.SetMemberRoleRequest) (*pb.SetMemberRoleResponse, error) {
	err := s.uc.SetMemberRole(ctx, req.GroupId, req.UserId, req.OperatorId, int32(req.Role))
	if err != nil {
		return nil, err
	}
	return &pb.SetMemberRoleResponse{Success: true}, nil
}

func (s *GroupService) KickMember(ctx context.Context, req *pb.KickMemberRequest) (*pb.KickMemberResponse, error) {
	err := s.uc.KickMember(ctx, req.GroupId, req.UserId, req.OperatorId)
	if err != nil {
		return nil, err
	}
	return &pb.KickMemberResponse{Success: true}, nil
}

func (s *GroupService) MuteMember(ctx context.Context, req *pb.MuteMemberRequest) (*pb.MuteMemberResponse, error) {
	err := s.uc.MuteMember(ctx, req.GroupId, req.UserId, req.OperatorId, req.DurationSeconds)
	if err != nil {
		return nil, err
	}
	return &pb.MuteMemberResponse{Success: true}, nil
}

func (s *GroupService) UnmuteMember(ctx context.Context, req *pb.UnmuteMemberRequest) (*pb.UnmuteMemberResponse, error) {
	err := s.uc.UnmuteMember(ctx, req.GroupId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.UnmuteMemberResponse{Success: true}, nil
}
