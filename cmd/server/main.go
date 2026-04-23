// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load file tree template
	template, err := config.LoadFileTreeTemplate(".")
	if err != nil {
		log.Fatalf("Failed to load file tree template: %v", err)
	}
	log.Printf("✓ File tree template loaded successfully (version: %s)", template.Version)

	// Initialize database
	db, err := database.NewDatabase(cfg.GetDSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := database.NewRedisClient(cfg.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize S3 service
	s3Service, err := storage.NewS3Service(
		cfg.AWS.Region,
		cfg.Storage.BucketName,
		cfg.AWS.UseIAMRole,
		cfg.AWS.AccessKeyID,
		cfg.AWS.SecretAccessKey,
		os.Getenv("AWS_ENDPOINT"),          // For MinIO (internal Docker access)
		os.Getenv("AWS_ENDPOINT_EXTERNAL"), // For presigned URLs (external client access)
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
		cfg.JWT.Issuer,
	)

	// Initialize repositories
	adminRepo := repository.NewAdminRepository(db.DB)
	tenantRepo := repository.NewTenantRepository(db.DB)
	fileRepo := repository.NewFileRepository(db.DB)
	auditRepo := repository.NewAuditRepository(db.DB, redisClient)
	tokenRepo := repository.NewTokenRepository(db.DB)
	syncTokenRepo := repository.NewSyncTokenRepository(db.DB)

	// Initialize services
	authService := service.NewAuthService(adminRepo, tenantRepo, tokenRepo, auditRepo, jwtService)
	adminService := service.NewAdminService(adminRepo, tenantRepo, tokenRepo, auditRepo, redisClient)
	tenantService := service.NewTenantService(tenantRepo, tokenRepo, auditRepo, redisClient)
	syncTokenService := service.NewSyncTokenService(syncTokenRepo, tenantRepo, auditRepo, redisClient)
	fileService := service.NewFileService(
		fileRepo,
		tenantRepo,
		auditRepo,
		s3Service,
		template,
		cfg.Storage.MaxFileSize,
		cfg.Storage.AllowedMimeTypes,
		cfg.Storage.PresignedExpiry,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(adminService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	syncTokenHandler := handlers.NewSyncTokenHandler(syncTokenService)
	fileHandler := handlers.NewFileHandler(fileService)
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	configHandler := handlers.NewConfigHandler(template)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, syncTokenService)
	rateLimiter := middleware.NewRateLimiter(10, 20) // 10 req/sec, burst 20

	// Setup router
	router := setupRouter(
		cfg,
		authHandler,
		adminHandler,
		tenantHandler,
		syncTokenHandler,
		fileHandler,
		healthHandler,
		configHandler,
		authMiddleware,
		rateLimiter,
	)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s (environment: %s)", cfg.Server.Port, cfg.Server.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

func setupRouter(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	adminHandler *handlers.AdminHandler,
	tenantHandler *handlers.TenantHandler,
	syncTokenHandler *handlers.SyncTokenHandler,
	fileHandler *handlers.FileHandler,
	healthHandler *handlers.HealthHandler,
	configHandler *handlers.ConfigHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1
	v1 := router.Group("/api/v1")
	v1.Use(rateLimiter.RateLimit())

	// Config routes (public - no auth required)
	config := v1.Group("/config")
	{
		config.GET("/template", configHandler.GetTemplate)
	}

	// Authentication routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/admin/login", authHandler.AdminLogin)
		auth.POST("/tenant/login", authHandler.TenantLogin)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	// Admin routes (requires admin authentication)
	admin := v1.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireAdmin())
	admin.Use(rateLimiter.RateLimitByUser())
	{
		// Password management
		admin.POST("/change_password", adminHandler.ChangePassword)

		// Tenant management
		admin.POST("/tenants", adminHandler.CreateTenant)
		admin.GET("/tenants", adminHandler.ListTenants)
		admin.GET("/tenants/:id", adminHandler.GetTenant)
		admin.POST("/tenants/:id/reset-password", adminHandler.ResetTenantPassword)
		admin.PATCH("/tenants/:id/status", adminHandler.UpdateTenantStatus)
		admin.DELETE("/tenants/:id", adminHandler.DeleteTenant)

		// Audit logs
		admin.GET("/audit-logs", adminHandler.GetAuditLogs)
		admin.POST("/audit-logs/rotate", adminHandler.RotateAuditLogs)

		// Sync token management
		admin.POST("/sync-tokens", syncTokenHandler.CreateSyncToken)
		admin.GET("/sync-tokens", syncTokenHandler.ListAllSyncTokens)
		admin.GET("/sync-tokens/:id", syncTokenHandler.GetSyncToken)
		admin.DELETE("/sync-tokens/cleanup", syncTokenHandler.CleanupRevokedTokens)
		admin.DELETE("/sync-tokens/:id/permanent", syncTokenHandler.DeleteSyncToken)
		admin.GET("/tenants/:id/sync-tokens", syncTokenHandler.ListTenantSyncTokens)
		admin.POST("/sync-tokens/:id/rotate", syncTokenHandler.RotateSyncToken)
		admin.DELETE("/sync-tokens/:id", syncTokenHandler.RevokeSyncToken)
		admin.GET("/sync-tokens/:id/stats", syncTokenHandler.GetSyncTokenStats)
	}

	// File routes (requires tenant authentication - accepts JWT OR sync tokens)
	files := v1.Group("/files")
	files.Use(authMiddleware.RequireAuthOrSync())
	files.Use(authMiddleware.RequireTenant())
	files.Use(authMiddleware.EnforceMustChangePassword()) // Enforce password change on first login
	files.Use(rateLimiter.RateLimitByUser())
	{
		files.POST("/upload", authMiddleware.RequirePermission("write"), fileHandler.ProxyUpload)
		files.POST("/upload-url", authMiddleware.RequirePermission("write"), fileHandler.GenerateUploadURL)
		files.POST("/complete-upload", authMiddleware.RequirePermission("write"), fileHandler.CompleteUpload)
		files.GET("", authMiddleware.RequirePermission("read"), fileHandler.ListFiles)
		files.GET("/:id/download-url", authMiddleware.RequirePermission("read"), fileHandler.GenerateDownloadURL)
		files.DELETE("/:id", authMiddleware.RequirePermission("delete"), fileHandler.DeleteFile)

		// Sync-specific endpoints (optimized for high-volume sync operations)
		files.GET("/sync/metadata", authMiddleware.RequirePermission("read"), fileHandler.GetSyncMetadata)
		files.POST("/sync/download-urls", authMiddleware.RequirePermission("read"), fileHandler.GenerateSyncDownloadURLs)
	}

	// Tenant routes (requires tenant authentication)
	tenant := v1.Group("/tenant")
	tenant.Use(authMiddleware.RequireAuth())
	tenant.Use(authMiddleware.RequireTenant())
	tenant.Use(authMiddleware.EnforceMustChangePassword()) // Enforce password change on first login
	{
		tenant.GET("/profile", tenantHandler.GetProfile)
		tenant.POST("/change-password", tenantHandler.ChangePassword)

		// Tenant's own sync tokens (view only)
		tenant.GET("/sync-tokens", syncTokenHandler.ListOwnSyncTokens)
		tenant.GET("/sync-tokens/:id", syncTokenHandler.GetOwnSyncToken)
		tenant.GET("/sync-tokens/:id/stats", syncTokenHandler.GetOwnSyncTokenStats)
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "route not found",
		})
	})

	return router
}
