// internal/handlers/file_handler.go
package handlers

import (
	"net/http"

	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) GenerateUploadURL(c *gin.Context) {
	var req models.UploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response, err := h.fileService.GenerateUploadURL(c.Request.Context(), tenantID, req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Upload URL generated", response)
}

func (h *FileHandler) CompleteUpload(c *gin.Context) {
	var req models.CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.fileService.CompleteUpload(c.Request.Context(), tenantID, req); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Upload completed", nil)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	files, err := h.fileService.ListFiles(c.Request.Context(), tenantID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Files retrieved", files)
}

func (h *FileHandler) GenerateDownloadURL(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response, err := h.fileService.GenerateDownloadURL(c.Request.Context(), tenantID, fileID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Download URL generated", response)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), tenantID, fileID); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "File deleted", nil)
}

// GetSyncMetadata returns file metadata for sync devices
func (h *FileHandler) GetSyncMetadata(c *gin.Context) {
	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response, err := h.fileService.GetSyncMetadata(c.Request.Context(), tenantID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync metadata retrieved", response)
}

// GenerateSyncDownloadURLs creates download URLs for multiple files (sync operation)
func (h *FileHandler) GenerateSyncDownloadURLs(c *gin.Context) {
	var req models.SyncDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrBadRequest)
		return
	}

	tenantID, err := middleware.GetUserID(c)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response, err := h.fileService.GenerateSyncDownloadURLs(c.Request.Context(), tenantID, req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sync download URLs generated", response)
}
