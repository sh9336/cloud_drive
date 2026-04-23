// internal/handlers/health_handler.go
package handlers

import (
	"context"
	"net/http"
	"time"

	"backend/internal/database"
	"backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db    *database.DB
	redis *database.RedisClient
}

func NewHealthHandler(db *database.DB, redis *database.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := "healthy"
	checks := make(map[string]interface{})

	// Check database
	if err := h.db.HealthCheck(ctx); err != nil {
		status = "unhealthy"
		checks["database"] = map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		checks["database"] = map[string]string{
			"status": "healthy",
		}
	}

	// Check Redis
	if err := h.redis.HealthCheck(ctx); err != nil {
		status = "unhealthy"
		checks["redis"] = map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		checks["redis"] = map[string]string{
			"status": "healthy",
		}
	}

	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"checks":    checks,
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.HealthCheck(ctx); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, err)
		return
	}

	if err := h.redis.HealthCheck(ctx); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}
