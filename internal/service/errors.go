package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrCodeExpired        = errors.New("sms code expired")
	ErrCodeInvalid        = errors.New("sms code invalid")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionRevoked     = errors.New("session revoked")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrHierarchyViolation = errors.New("hierarchy violation")
	ErrConflict           = errors.New("resource conflict")
	ErrInvalidInput       = errors.New("invalid input")
)
