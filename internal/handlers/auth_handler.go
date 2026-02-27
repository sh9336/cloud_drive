// internal/handlers/auth_handler.go
package handlers

import (
	"net/http"

	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	response, err := h.authService.AdminLogin(c.Request.Context(), req, ip, userAgent)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

func (h *AuthHandler) TenantLogin(c *gin.Context) {
	var req models.TenantLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	response, err := h.authService.TenantLogin(c.Request.Context(), req, ip, userAgent)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed", response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}
