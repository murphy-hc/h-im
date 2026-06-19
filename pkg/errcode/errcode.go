// Package errcode maps protocol-level error codes to Go errors.
package errcode

import (
	"errors"
	"fmt"
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
