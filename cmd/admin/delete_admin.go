// cmd/admin/delete_admin.go
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
	"backend/internal/models"
	"backend/internal/repository"
)

func main() {
	fmt.Println("=== Delete Super Admin ===\n")

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
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	// Get email
	fmt.Print("Admin Email to delete: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Find admin
	admin, err := repo.GetByEmail(ctx, email)
	if err != nil {
		log.Fatalf("Admin not found: %v", err)
	}

	fmt.Printf("\nFound: %s (%s)\n", admin.FullName, admin.Email)
	fmt.Print("\n⚠️  Are you sure you want to delete this admin? (yes/no): ")

	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "yes" {
		fmt.Println("Deletion cancelled.")
		return
	}

	// Delete admin
	if err := repo.Delete(ctx, admin.ID); err != nil {
		log.Fatalf("Failed to delete admin: %v", err)
	}

	// Create audit log
	_ = auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "system",
		Action:       "admin_deleted",
		ResourceType: stringPtr("admin"),
		ResourceID:   &admin.ID,
		Metadata: map[string]interface{}{
			"deleted_via":   "cli",
			"deleted_email": email,
		},
		Status: "success",
	})

	fmt.Println("\n✅ Admin deleted successfully!")
}

func stringPtr(s string) *string {
	return &s
}
