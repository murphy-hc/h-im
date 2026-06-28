// Package errcode maps protocol-level error codes to Go errors.
package errcode

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Domain error codes, kept in sync with proto/him/common/v1/error.proto.
const (
	// Auth 1000-1099
	TokenExpired     = 1000
	TokenInvalid     = 1001
	PermissionDenied = 1002

	// User 1100-1199
	UserNotFound     = 1100
	UserAlreadyExists = 1101

	// Message 1200-1299
	MessageNotFound  = 1200
	MessageSendFailed = 1201

	// Contact 1300-1399
	AlreadyFriends       = 1300
	NotFriends           = 1301
	FriendRequestExpired = 1302

	// Group 1400-1499
	GroupNotFound         = 1400
	NotGroupMember        = 1401
	GroupMemberLimit      = 1402
	GroupPermissionDenied = 1403

	// Chatroom 1500-1599
	RoomNotFound     = 1500
	NotRoomMember    = 1501
	InvalidPriority  = 1502

	// Media 1600-1699
	MediaUploadFailed = 1600
	MediaNotFound     = 1601
	MediaSizeExceeded = 1602
)

// Error wraps a domain error code with a human-readable message.
type Error struct {
	Code    int32
	Message string
	Details []string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New returns an *Error with the given code and message.
func New(code int32, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

// Newf returns an *Error with a formatted message.
func Newf(code int32, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// WithDetails appends detail strings to the error.
func (e *Error) WithDetails(d ...string) *Error {
	e.Details = append(e.Details, d...)
	return e
}

// IsDomainError checks whether err is an *Error.
func IsDomainError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

// Code extracts the domain code from an error, returning 0 if not a domain error.
func Code(err error) int32 {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 0
}

// GRPCCode maps a domain error code to a gRPC status code.
func GRPCCode(code int32) codes.Code {
	switch {
	// Auth
	case code >= 1000 && code <= 1001:
		return codes.Unauthenticated
	case code == 1002:
		return codes.PermissionDenied

	// User
	case code == 1100:
		return codes.NotFound
	case code == 1101:
		return codes.AlreadyExists

	// Message
	case code == 1200:
		return codes.NotFound
	case code == 1201:
		return codes.Internal

	// Contact
	case code == 1300:
		return codes.AlreadyExists
	case code == 1301:
		return codes.FailedPrecondition
	case code == 1302:
		return codes.FailedPrecondition

	// Group
	case code == 1400:
		return codes.NotFound
	case code == 1401:
		return codes.FailedPrecondition
	case code == 1402:
		return codes.ResourceExhausted
	case code == 1403:
		return codes.PermissionDenied

	// Chatroom
	case code == 1500:
		return codes.NotFound
	case code == 1501:
		return codes.FailedPrecondition
	case code == 1502:
		return codes.InvalidArgument

	// Media
	case code == 1600:
		return codes.Internal
	case code == 1601:
		return codes.NotFound
	case code == 1602:
		return codes.InvalidArgument

	default:
		return codes.Internal
	}
}

// ToGRPCStatus converts an error to a gRPC status. If the error is a domain
// *Error, the mapped gRPC code is used; otherwise codes.Internal is returned.
func ToGRPCStatus(err error) *status.Status {
	var e *Error
	if errors.As(err, &e) {
		return status.New(GRPCCode(e.Code), e.Message)
	}
	return status.New(codes.Internal, err.Error())
}
