// cmd/admin/enable_admin.go
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
	fmt.Println("=== Enable/Disable Super Admin ===\n")

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
	fmt.Print("Admin Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Find admin
	admin, err := repo.GetByEmail(ctx, email)
	if err != nil {
		log.Fatalf("Admin not found: %v", err)
	}

	currentStatus := "Active"
	if !admin.IsActive {
		currentStatus = "Disabled"
	}

	fmt.Printf("\nFound: %s (%s)\n", admin.FullName, admin.Email)
	fmt.Printf("Current Status: %s\n\n", currentStatus)

	fmt.Print("Action (enable/disable): ")
	action, _ := reader.ReadString('\n')
	action = strings.TrimSpace(strings.ToLower(action))

	var isActive bool
	var reason *string

	switch action {
	case "enable":
		isActive = true
	case "disable":
		isActive = false
		fmt.Print("Reason for disabling: ")
		reasonInput, _ := reader.ReadString('\n')
		reasonInput = strings.TrimSpace(reasonInput)
		reason = &reasonInput
	default:
		log.Fatal("Invalid action. Use 'enable' or 'disable'")
	}

	// Update status
	if err := repo.UpdateStatus(ctx, admin.ID, isActive, reason); err != nil {
		log.Fatalf("Failed to update status: %v", err)
	}

	// Create audit log
	_ = auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "system",
		Action:       fmt.Sprintf("admin_%sd", action),
		ResourceType: stringPtr("admin"),
		ResourceID:   &admin.ID,
		Metadata: map[string]interface{}{
			"action_via": "cli",
			"reason":     reason,
		},
		Status: "success",
	})

	fmt.Printf("\n✅ Admin %sd successfully!\n", action)
}

func stringPtr(s string) *string {
	return &s
}
