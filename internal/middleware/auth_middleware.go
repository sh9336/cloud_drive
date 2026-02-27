// internal/middleware/auth_middleware.go
package middleware

import (
	"net/http"
	"strings"

	"backend/internal/auth"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtService       *auth.JWTService
	syncTokenService *service.SyncTokenService
}

func NewAuthMiddleware(jwtService *auth.JWTService, syncTokenService *service.SyncTokenService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:       jwtService,
		syncTokenService: syncTokenService,
	}
}

// RequireAuth validates JWT and sets user info in context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrInvalidToken)
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_type", claims.UserType)
		c.Set("auth_type", "jwt")

		c.Next()
	}
}

// RequireSyncAuth validates sync token and sets sync info in context
func (m *AuthMiddleware) RequireSyncAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if it's a sync token (starts with "sync_")
		if !strings.HasPrefix(token, "sync_") {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrInvalidSyncToken)
			c.Abort()
			return
		}

		claims, err := m.syncTokenService.ValidateSyncToken(c.Request.Context(), token)
		if err != nil {
			utils.HandleError(c, err)
			c.Abort()
			return
		}

		// Set sync token info in context
		c.Set("sync_token_id", claims.TokenID)
		c.Set("user_id", claims.TenantID) // Tenant ID
		c.Set("user_type", "tenant")      // Sync tokens belong to tenants
		c.Set("auth_type", "sync")
		c.Set("sync_can_read", claims.CanRead)
		c.Set("sync_can_write", claims.CanWrite)
		c.Set("sync_can_delete", claims.CanDelete)

		c.Next()
	}
}

// RequireAuthOrSync accepts both JWT and sync tokens
func (m *AuthMiddleware) RequireAuthOrSync() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if it's a sync token
		if strings.HasPrefix(token, "sync_") {
			// Validate as sync token
			claims, err := m.syncTokenService.ValidateSyncToken(c.Request.Context(), token)
			if err != nil {
				utils.HandleError(c, err)
				c.Abort()
				return
			}

			// Set sync token info in context
			c.Set("sync_token_id", claims.TokenID)
			c.Set("user_id", claims.TenantID)
			c.Set("user_type", "tenant")
			c.Set("auth_type", "sync")
			c.Set("sync_can_read", claims.CanRead)
			c.Set("sync_can_write", claims.CanWrite)
			c.Set("sync_can_delete", claims.CanDelete)
		} else {
			// Validate as JWT
			claims, err := m.jwtService.ValidateAccessToken(token)
			if err != nil {
				utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrInvalidToken)
				c.Abort()
				return
			}

			// Set user info in context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_type", claims.UserType)
			c.Set("auth_type", "jwt")
		}

		c.Next()
	}
}

// RequireAdmin ensures user is a super admin
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, utils.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireTenant ensures user is a tenant
func (m *AuthMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "tenant" {
			utils.ErrorResponse(c, http.StatusForbidden, utils.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermission checks sync token permissions
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, _ := c.Get("auth_type")

		// JWT users have all permissions
		if authType == "jwt" {
			c.Next()
			return
		}

		// Check sync token permissions
		if authType == "sync" {
			switch permission {
			case "read":
				canRead, _ := c.Get("sync_can_read")
				if !canRead.(bool) {
					utils.ErrorResponse(c, http.StatusForbidden, utils.ErrInsufficientPermissions)
					c.Abort()
					return
				}
			case "write":
				canWrite, _ := c.Get("sync_can_write")
				if !canWrite.(bool) {
					utils.ErrorResponse(c, http.StatusForbidden, utils.ErrInsufficientPermissions)
					c.Abort()
					return
				}
			case "delete":
				canDelete, _ := c.Get("sync_can_delete")
				if !canDelete.(bool) {
					utils.ErrorResponse(c, http.StatusForbidden, utils.ErrInsufficientPermissions)
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// Check Authorization header
	bearerToken := c.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, utils.ErrUnauthorized
	}
	return userID.(uuid.UUID), nil
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) (string, error) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", utils.ErrUnauthorized
	}
	return email.(string), nil
}

// GetUserType extracts user type from context
func GetUserType(c *gin.Context) (string, error) {
	userType, exists := c.Get("user_type")
	if !exists {
		return "", utils.ErrUnauthorized
	}
	return userType.(string), nil
}

// GetAuthType extracts auth type (jwt or sync) from context
func GetAuthType(c *gin.Context) string {
	authType, exists := c.Get("auth_type")
	if !exists {
		return ""
	}
	return authType.(string)
}

// GetSyncTokenID extracts sync token ID from context
func GetSyncTokenID(c *gin.Context) (uuid.UUID, error) {
	tokenID, exists := c.Get("sync_token_id")
	if !exists {
		return uuid.Nil, utils.ErrUnauthorized
	}
	return tokenID.(uuid.UUID), nil
}

// EnforceMustChangePassword ensures tenants change password on first login
// Allows only: change-password, profile, logout endpoints
func (m *AuthMiddleware) EnforceMustChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, _ := c.Get("user_type")

		// Only applies to tenants
		if userType != "tenant" {
			c.Next()
			return
		}

		// Extract JWT token and check for MustChangePassword claim
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Check if password must be changed
		if !claims.MustChangePassword {
			c.Next()
			return
		}

		// Whitelist allowed endpoints when password must be changed
		path := c.Request.URL.Path
		method := c.Request.Method

		// Allowed endpoints:
		// POST /api/v1/tenant/change-password
		// GET /api/v1/tenant/profile
		// POST /api/v1/auth/logout
		allowedEndpoints := map[string]bool{
			"POST:/api/v1/tenant/change-password": true,
			"GET:/api/v1/tenant/profile":          true,
			"POST:/api/v1/auth/logout":            true,
		}

		endpointKey := method + ":" + path
		if !allowedEndpoints[endpointKey] {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Password change required. Please call POST /api/v1/tenant/change-password first",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
