// internal/utils/response.go
package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, code int, err error) {
	c.JSON(code, Response{
		Success: false,
		Error:   err.Error(),
	})
}

func HandleError(c *gin.Context, err error) {
	log.Printf("Error handled: %v", err)
	switch err {
	case ErrUnauthorized, ErrInvalidCredentials, ErrInvalidToken, ErrTokenExpired, ErrTokenRevoked:
		ErrorResponse(c, http.StatusUnauthorized, err)
	case ErrForbidden, ErrAccountDisabled:
		ErrorResponse(c, http.StatusForbidden, err)
	case ErrNotFound:
		ErrorResponse(c, http.StatusNotFound, err)
	case ErrBadRequest, ErrEmailAlreadyExists, ErrFileTooLarge, ErrInvalidMimeType:
		ErrorResponse(c, http.StatusBadRequest, err)
	case ErrMustChangePassword, ErrPasswordExpired:
		ErrorResponse(c, http.StatusPreconditionRequired, err)
	default:
		ErrorResponse(c, http.StatusInternalServerError, ErrInternalServer)
	}
}
