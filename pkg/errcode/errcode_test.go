package errcode

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestNewAndCode(t *testing.T) {
	err := New(UserNotFound, "user not found")
	if err.Code != UserNotFound {
		t.Fatalf("expected code %d, got %d", UserNotFound, err.Code)
	}
	if !IsDomainError(err) {
		t.Fatal("expected IsDomainError to return true")
	}
	if Code(err) != UserNotFound {
		t.Fatalf("expected Code %d, got %d", UserNotFound, Code(err))
	}
}

func TestNewf(t *testing.T) {
	err := Newf(MessageSendFailed, "send to %s failed", "user-1")
	if err.Code != MessageSendFailed {
		t.Fatalf("expected code %d, got %d", MessageSendFailed, err.Code)
	}
}

func TestIsDomainErrorRegular(t *testing.T) {
	if IsDomainError(errors.New("regular error")) {
		t.Fatal("regular error should not be domain error")
	}
	if Code(errors.New("regular")) != 0 {
		t.Fatal("regular error should return code 0")
	}
}

func TestGRPCCodeMapping(t *testing.T) {
	tests := []struct {
		code     int32
		expected codes.Code
	}{
		{TokenExpired, codes.Unauthenticated},
		{TokenInvalid, codes.Unauthenticated},
		{PermissionDenied, codes.PermissionDenied},
		{UserNotFound, codes.NotFound},
		{UserAlreadyExists, codes.AlreadyExists},
		{MessageNotFound, codes.NotFound},
		{MessageSendFailed, codes.Internal},
		{AlreadyFriends, codes.AlreadyExists},
		{NotFriends, codes.FailedPrecondition},
		{FriendRequestExpired, codes.FailedPrecondition},
		{GroupNotFound, codes.NotFound},
		{NotGroupMember, codes.FailedPrecondition},
		{GroupMemberLimit, codes.ResourceExhausted},
		{GroupPermissionDenied, codes.PermissionDenied},
		{RoomNotFound, codes.NotFound},
		{NotRoomMember, codes.FailedPrecondition},
		{InvalidPriority, codes.InvalidArgument},
		{MediaUploadFailed, codes.Internal},
		{MediaNotFound, codes.NotFound},
		{MediaSizeExceeded, codes.InvalidArgument},
	}

	for _, tt := range tests {
		if got := GRPCCode(tt.code); got != tt.expected {
			t.Errorf("GRPCCode(%d) = %v, want %v", tt.code, got, tt.expected)
		}
	}
}

func TestToGRPCStatus(t *testing.T) {
	err := New(UserNotFound, "user not found")
	st := ToGRPCStatus(err)
	if st.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", st.Code())
	}
	if st.Message() != "user not found" {
		t.Fatalf("expected 'user not found', got %q", st.Message())
	}
}

func TestToGRPCStatusRegular(t *testing.T) {
	st := ToGRPCStatus(errors.New("something broke"))
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal for regular error, got %v", st.Code())
	}
}
