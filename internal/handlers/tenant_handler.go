// internal/handlers/tenant_handler.go
package handlers

import (
	"net/http"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService *service.TenantService
}

func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

func (h *TenantHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.tenantService.ChangePassword(c.Request.Context(), tenantID, req); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

func (h *TenantHandler) GetProfile(c *gin.Context) {
	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	tenant, err := h.tenantService.GetProfile(c.Request.Context(), tenantID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved", tenant)
}
