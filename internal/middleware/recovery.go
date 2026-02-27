// internal/middleware/recovery.go
package middleware

import (
	"log"
	"net/http"

	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer)
				c.Abort()
			}
		}()
		c.Next()
	}
}
