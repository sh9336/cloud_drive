// internal/handlers/sync_token_handler.go
package handlers

import (
	"fmt"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SyncTokenHandler struct {
	syncTokenService *service.SyncTokenService
}

func NewSyncTokenHandler(syncTokenService *service.SyncTokenService) *SyncTokenHandler {
	return &SyncTokenHandler{
		syncTokenService: syncTokenService,
	}
}

// CreateSyncToken creates a new sync token (admin only)
func (h *SyncTokenHandler) CreateSyncToken(c *gin.Context) {
	var req models.CreateSyncTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response, err := h.syncTokenService.CreateSyncToken(c.Request.Context(), req, adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Sync token created successfully", response)
}

// ListAllSyncTokens lists all sync tokens (admin only)
func (h *SyncTokenHandler) ListAllSyncTokens(c *gin.Context) {
	tokens, err := h.syncTokenService.ListAllSyncTokens(c.Request.Context())
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync tokens retrieved", tokens)
}

// ListTenantSyncTokens lists sync tokens for a specific tenant (admin only)
func (h *SyncTokenHandler) ListTenantSyncTokens(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tokens, err := h.syncTokenService.ListTenantSyncTokens(c.Request.Context(), tenantID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync tokens retrieved", tokens)
}

// ListOwnSyncTokens lists sync tokens for the authenticated tenant
func (h *SyncTokenHandler) ListOwnSyncTokens(c *gin.Context) {
	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	tokens, err := h.syncTokenService.ListTenantSyncTokens(c.Request.Context(), tenantID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync tokens retrieved", tokens)
}

// GetSyncToken gets a sync token by ID (admin only)
func (h *SyncTokenHandler) GetSyncToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	token, err := h.syncTokenService.GetSyncToken(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token retrieved", token)
}

// GetOwnSyncToken gets own sync token by ID (tenant only)
func (h *SyncTokenHandler) GetOwnSyncToken(c *gin.Context) {
	idStr := c.Param("id")
	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	token, err := h.syncTokenService.GetSyncToken(c.Request.Context(), tokenID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Verify token belongs to this tenant
	if token.TenantID != tenantID {
		utils.ErrorResponse(c, http.StatusForbidden, utils.ErrForbidden)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token retrieved", token)
}

// RotateSyncToken rotates a sync token (admin only)
func (h *SyncTokenHandler) RotateSyncToken(c *gin.Context) {
	idStr := c.Param("id")
	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	var req models.RotateSyncTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response, err := h.syncTokenService.RotateSyncToken(c.Request.Context(), tokenID, adminID, req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token rotated successfully", response)
}

// RevokeSyncToken revokes a sync token (admin only)
func (h *SyncTokenHandler) RevokeSyncToken(c *gin.Context) {
	idStr := c.Param("id")
	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	var req models.RevokeSyncTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.syncTokenService.RevokeSyncToken(c.Request.Context(), tokenID, adminID, req.Reason); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token revoked successfully", nil)
}

// GetSyncTokenStats gets statistics for a sync token (admin only)
func (h *SyncTokenHandler) GetSyncTokenStats(c *gin.Context) {
	idStr := c.Param("id")
	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	stats, err := h.syncTokenService.GetSyncTokenStats(c.Request.Context(), tokenID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token statistics retrieved", stats)
}

// GetOwnSyncTokenStats gets statistics for own sync token (tenant only)
func (h *SyncTokenHandler) GetOwnSyncTokenStats(c *gin.Context) {
	idStr := c.Param("id")
	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// First verify token belongs to this tenant
	token, err := h.syncTokenService.GetSyncToken(c.Request.Context(), tokenID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if token.TenantID != tenantID {
		utils.ErrorResponse(c, http.StatusForbidden, utils.ErrForbidden)
		return
	}

	stats, err := h.syncTokenService.GetSyncTokenStats(c.Request.Context(), tokenID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token statistics retrieved", stats)
}

// CleanupRevokedTokens deletes all revoked/deactivated sync tokens (admin only)
func (h *SyncTokenHandler) CleanupRevokedTokens(c *gin.Context) {
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	count, err := h.syncTokenService.CleanupRevokedTokens(c.Request.Context(), adminID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, fmt.Sprintf("Cleanup successful, %d tokens deleted", count), gin.H{
		"deleted_count": count,
	})
}

// DeleteSyncToken permanently deletes an individual sync token (admin only)
func (h *SyncTokenHandler) DeleteSyncToken(c *gin.Context) {
	idStr := c.Param("id")
	tokenID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.syncTokenService.DeleteSyncToken(c.Request.Context(), tokenID, adminID); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync token permanently deleted", nil)
}
