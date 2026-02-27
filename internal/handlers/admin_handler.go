// internal/handlers/admin_handler.go
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

func (h *AdminHandler) CreateTenant(c *gin.Context) {
	var req models.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	tenant, tempPassword, err := h.adminService.CreateTenant(c.Request.Context(), req, adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Tenant created successfully", gin.H{
		"tenant":             tenant,
		"temporary_password": tempPassword,
	})
}

func (h *AdminHandler) ListTenants(c *gin.Context) {
	tenants, err := h.adminService.ListTenants(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tenants retrieved", tenants)
}

func (h *AdminHandler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenant, err := h.adminService.GetTenant(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tenant retrieved", tenant)
}

func (h *AdminHandler) ResetTenantPassword(c *gin.Context) {
	idStr := c.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	tempPassword, err := h.adminService.ResetTenantPassword(c.Request.Context(), tenantID, adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password reset successful", gin.H{
		"temporary_password": tempPassword,
	})
}

func (h *AdminHandler) UpdateTenantStatus(c *gin.Context) {
	idStr := c.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	var req models.UpdateTenantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log bind error for debugging and return a clearer message
		log.Printf("UpdateTenantStatus: bind error: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.IsActive == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("is_active is required"))
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.adminService.UpdateTenantStatus(c.Request.Context(), tenantID, adminID, req); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tenant status updated", nil)
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.adminService.ChangePassword(c.Request.Context(), adminID, req); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := models.AuditLogFilter{
		ActorType:  c.Query("actor_type"),
		ActorEmail: c.Query("actor_email"),
		Action:     c.Query("action"),
		Status:     c.Query("status"),
		Page:       page,
		Limit:      limit,
	}

	if startStr := c.Query("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndDate = &t
		}
	}

	response, err := h.adminService.GetAuditLogs(c.Request.Context(), filter)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Audit logs retrieved", response)
}

func (h *AdminHandler) RotateAuditLogs(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid days parameter"))
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	count, err := h.adminService.RotateAuditLogs(c.Request.Context(), days, adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, fmt.Sprintf("Audit logs rotated, %d records deleted", count), gin.H{
		"deleted_count":  count,
		"retention_days": days,
	})
}

func (h *AdminHandler) DeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.adminService.DeleteTenant(c.Request.Context(), tenantID, adminID); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tenant permanently deleted", nil)
}
