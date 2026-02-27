// internal/models/auth.go
package models

import "github.com/google/uuid"

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"` // seconds
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	FullName           string    `json:"full_name"`
	UserType           string    `json:"user_type"` // 'admin' or 'tenant'
	MustChangePassword bool      `json:"must_change_password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
