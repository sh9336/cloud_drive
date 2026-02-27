// cmd/admin/rotate_sync_token.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("=== Rotate Sync Token ===\n")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.GetDSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	syncTokenRepo := repository.NewSyncTokenRepository(db.DB)
	tenantRepo := repository.NewTenantRepository(db.DB)
	auditRepo := repository.NewAuditRepository(db.DB)
	adminRepo := repository.NewAdminRepository(db.DB)

	syncTokenService := service.NewSyncTokenService(syncTokenRepo, tenantRepo, auditRepo)

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	// Get admin email
	fmt.Print("Your admin email: ")
	adminEmail, _ := reader.ReadString('\n')
	adminEmail = strings.TrimSpace(adminEmail)

	admin, err := adminRepo.GetByEmail(ctx, adminEmail)
	if err != nil {
		log.Fatalf("Admin not found: %v", err)
	}

	// Get token ID
	fmt.Print("Sync Token ID: ")
	tokenIDStr, _ := reader.ReadString('\n')
	tokenIDStr = strings.TrimSpace(tokenIDStr)

	tokenID, err := uuid.Parse(tokenIDStr)
	if err != nil {
		log.Fatal("Invalid token ID")
	}

	// Get token details
	token, err := syncTokenService.GetSyncToken(ctx, tokenID)
	if err != nil {
		log.Fatalf("Token not found: %v", err)
	}

	fmt.Printf("\nToken: %s\n", token.Name)
	fmt.Printf("Current expires: %s\n\n", token.ExpiresAt.Format("2006-01-02"))

	// Get new expiry
	fmt.Print("New expiry in days (1-365): ")
	expiryInput, _ := reader.ReadString('\n')
	expiryDays, _ := strconv.Atoi(strings.TrimSpace(expiryInput))

	if expiryDays < 1 || expiryDays > 365 {
		log.Fatal("Expiry days must be between 1 and 365")
	}

	// Get grace period
	fmt.Print("Grace period in days (0-30, default 7): ")
	graceInput, _ := reader.ReadString('\n')
	graceDays, _ := strconv.Atoi(strings.TrimSpace(graceInput))

	if graceDays < 0 || graceDays > 30 {
		graceDays = 7
	}

	// Rotate token
	req := models.RotateSyncTokenRequest{
		ExpiresInDays:   expiryDays,
		GracePeriodDays: graceDays,
	}

	response, err := syncTokenService.RotateSyncToken(ctx, tokenID, admin.ID, req)
	if err != nil {
		log.Fatalf("Failed to rotate token: %v", err)
	}

	fmt.Println("\n✅ Sync token rotated successfully!")
	fmt.Println("\n⚠️  IMPORTANT: Save this new token securely!")
	fmt.Printf("\nNew Token: %s\n", response.NewToken)
	fmt.Printf("\nOld Token ID: %s\n", response.OldTokenID)
	fmt.Printf("New Token ID: %s\n", response.NewTokenID)
	fmt.Printf("Expires: %s\n", response.ExpiresAt.Format("2006-01-02"))
	fmt.Printf("\nGrace Period: %d days (old token works until %s)\n",
		response.GracePeriodDays,
		response.GraceEndsAt.Format("2006-01-02"))
}
