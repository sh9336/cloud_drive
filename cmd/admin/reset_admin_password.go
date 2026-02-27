// cmd/admin/reset_admin_password.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"
)

func main() {
	fmt.Println("=== Reset Super Admin Password ===\n")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.GetDSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewAdminRepository(db.DB)
	auditRepo := repository.NewAuditRepository(db.DB)
	tokenRepo := repository.NewTokenRepository(db.DB)
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	// Get email
	fmt.Print("Admin Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Find admin
	admin, err := repo.GetByEmail(ctx, email)
	if err != nil {
		log.Fatalf("Admin not found: %v", err)
	}

	fmt.Printf("Found: %s (%s)\n\n", admin.FullName, admin.Email)

	// Generate temporary password
	tempPassword, err := auth.GenerateTemporaryPassword()
	if err != nil {
		log.Fatalf("Failed to generate password: %v", err)
	}

	// Hash password
	passwordHash, err := auth.HashPassword(tempPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Update password
	if err := repo.UpdatePassword(ctx, admin.ID, passwordHash); err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	// Revoke all existing tokens
	_ = tokenRepo.RevokeAllForUser(ctx, "admin", admin.ID, "password_reset")

	// Create audit log
	_ = auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "system",
		Action:       "password_reset",
		ResourceType: stringPtr("admin"),
		ResourceID:   &admin.ID,
		Metadata: map[string]interface{}{
			"reset_via": "cli",
		},
		Status: "success",
	})

	fmt.Println("✅ Password reset successfully!")
	fmt.Printf("\nTemporary Password: %s\n", tempPassword)
	fmt.Println("\n⚠️  IMPORTANT:")
	fmt.Println("   - Write this down immediately")
	fmt.Println("   - User must change password on next login")
	fmt.Println("   - All existing sessions have been invalidated")
}

func stringPtr(s string) *string {
	return &s
}
