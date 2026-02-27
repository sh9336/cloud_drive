// internal/handlers/config_handler.go
package handlers

import (
	"net/http"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	template *models.FileTreeTemplate
}

func NewConfigHandler(template *models.FileTreeTemplate) *ConfigHandler {
	return &ConfigHandler{
		template: template,
	}
}

// GetTemplate returns the current file tree template
// GET /api/v1/config/template
func (h *ConfigHandler) GetTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"template":    h.template,
		"version":     h.template.Version,
		"timestamp":   time.Now().UTC(),
		"root_path":   h.template.RootPath,
		"nodes_count": len(h.template.Nodes),
	})
}
