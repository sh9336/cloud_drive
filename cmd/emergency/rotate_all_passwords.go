// cmd/emergency/rotate_all_passwords.go
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
	"backend/internal/repository"
)

func main() {
	fmt.Println("=== EMERGENCY: Rotate All Passwords ===\n")
	fmt.Println("⚠️  WARNING: This will reset passwords for ALL admins and tenants!")
	fmt.Println("⚠️  All users will need to reset their passwords on next login.")
	fmt.Print("\nAre you absolutely sure? (type 'CONFIRM' to proceed): ")

	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "CONFIRM" {
		fmt.Println("Operation cancelled.")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.GetDSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	adminRepo := repository.NewAdminRepository(db.DB)
	tenantRepo := repository.NewTenantRepository(db.DB)
	tokenRepo := repository.NewTokenRepository(db.DB)
	ctx := context.Background()

	fmt.Println("\nProcessing admins...")
	admins, _ := adminRepo.List(ctx)
	for _, admin := range admins {
		tempPassword, _ := auth.GenerateTemporaryPassword()
		passwordHash, _ := auth.HashPassword(tempPassword)
		_ = adminRepo.UpdatePassword(ctx, admin.ID, passwordHash)
		fmt.Printf("  ✓ %s: %s\n", admin.Email, tempPassword)
	}

	fmt.Println("\nProcessing tenants...")
	tenants, _ := tenantRepo.List(ctx)
	for _, tenant := range tenants {
		tempPassword, _ := auth.GenerateTemporaryPassword()
		passwordHash, _ := auth.HashPassword(tempPassword)
		_ = tenantRepo.UpdatePassword(ctx, tenant.ID, passwordHash)
		fmt.Printf("  ✓ %s: %s\n", tenant.Email, tempPassword)
	}

	// Revoke all tokens
	fmt.Println("\nRevoking all sessions...")
	for _, admin := range admins {
		_ = tokenRepo.RevokeAllForUser(ctx, "admin", admin.ID, "emergency_rotation")
	}
	for _, tenant := range tenants {
		_ = tokenRepo.RevokeAllForUser(ctx, "tenant", tenant.ID, "emergency_rotation")
	}

	fmt.Println("\n✅ Emergency password rotation complete!")
	fmt.Println("\n⚠️  SAVE THESE PASSWORDS SECURELY!")
	fmt.Println("   All users must change their password on next login.")
}
