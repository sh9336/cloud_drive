// internal/utils/errors.go
package utils

import "errors"

var (
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrNotFound                = errors.New("not found")
	ErrBadRequest              = errors.New("bad request")
	ErrInternalServer          = errors.New("internal server error")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrAccountDisabled         = errors.New("account is disabled")
	ErrMustChangePassword      = errors.New("password must be changed")
	ErrPasswordExpired         = errors.New("password has expired")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token has expired")
	ErrTokenRevoked            = errors.New("token has been revoked")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrFileTooLarge            = errors.New("file size exceeds maximum allowed")
	ErrInvalidMimeType         = errors.New("file type not allowed")
	ErrInvalidSyncToken        = errors.New("invalid sync token")
	ErrSyncTokenExpired        = errors.New("sync token has expired")
	ErrSyncTokenRevoked        = errors.New("sync token has been revoked")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

func StringPtr(s string) *string {
	return &s
}
