// cmd/admin/revoke_sync_token.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("=== Revoke Sync Token ===\n")

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
	fmt.Print("Sync Token ID to revoke: ")
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
	fmt.Printf("Tenant ID: %s\n\n", token.TenantID)

	fmt.Print("⚠️  Are you sure you want to revoke this token? (yes/no): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "yes" {
		fmt.Println("Revocation cancelled.")
		return
	}

	// Get reason
	fmt.Print("Reason for revocation: ")
	reason, _ := reader.ReadString('\n')
	reason = strings.TrimSpace(reason)

	if reason == "" {
		log.Fatal("Reason is required")
	}

	// Revoke token
	if err := syncTokenService.RevokeSyncToken(ctx, tokenID, admin.ID, reason); err != nil {
		log.Fatalf("Failed to revoke token: %v", err)
	}

	fmt.Println("\n✅ Sync token revoked successfully!")
}
