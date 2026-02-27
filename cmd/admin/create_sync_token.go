// cmd/admin/create_sync_token.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/service"

	"golang.org/x/term"
)

func main() {
	fmt.Println("=== Create Sync Token ===\n")

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
	fmt.Print("Admin email: ")
	adminEmail, _ := reader.ReadString('\n')
	adminEmail = strings.TrimSpace(adminEmail)

	if adminEmail == "" {
		log.Fatal("Admin email cannot be empty")
	}

	admin, err := adminRepo.GetByEmail(ctx, adminEmail)
	if err != nil {
		log.Fatalf("Admin not found: %v", err)
	}

	// Verify admin password
	fmt.Print("Admin password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	fmt.Println() // Print newline after password input

	err = auth.CheckPassword(string(password), admin.PasswordHash)
	if err != nil {
		log.Fatal("❌ Invalid password. Access denied.")
	}

	fmt.Println("✅ Password verified!\n")

	// List tenants
	fmt.Println("\n--- Available Tenants ---")
	tenants, _ := tenantRepo.List(ctx)
	for i, t := range tenants {
		fmt.Printf("%d. %s (%s)\n", i+1, t.Email, t.FullName)
	}

	// Get tenant selection
	fmt.Print("\nSelect tenant number: ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)
	tenantIndex, _ := strconv.Atoi(selection)

	if tenantIndex < 1 || tenantIndex > len(tenants) {
		log.Fatal("Invalid tenant selection")
	}

	selectedTenant := tenants[tenantIndex-1]

	// Get token name
	fmt.Print("\nToken name (e.g., 'Production Server'): ")
	tokenName, _ := reader.ReadString('\n')
	tokenName = strings.TrimSpace(tokenName)

	// Get permissions
	fmt.Print("Can read? (y/n): ")
	canReadInput, _ := reader.ReadString('\n')
	canRead := strings.ToLower(strings.TrimSpace(canReadInput)) == "y"

	fmt.Print("Can write? (y/n): ")
	canWriteInput, _ := reader.ReadString('\n')
	canWrite := strings.ToLower(strings.TrimSpace(canWriteInput)) == "y"

	fmt.Print("Can delete? (y/n): ")
	canDeleteInput, _ := reader.ReadString('\n')
	canDelete := strings.ToLower(strings.TrimSpace(canDeleteInput)) == "y"

	// Get expiry
	fmt.Print("Expires in how many days? (1-365): ")
	expiryInput, _ := reader.ReadString('\n')
	expiryDays, _ := strconv.Atoi(strings.TrimSpace(expiryInput))

	if expiryDays < 1 || expiryDays > 365 {
		log.Fatal("Expiry days must be between 1 and 365")
	}

	// Create sync token
	req := models.CreateSyncTokenRequest{
		TenantID:      selectedTenant.ID,
		Name:          tokenName,
		CanRead:       canRead,
		CanWrite:      canWrite,
		CanDelete:     canDelete,
		ExpiresInDays: expiryDays,
	}

	response, err := syncTokenService.CreateSyncToken(ctx, req, admin.ID)
	if err != nil {
		log.Fatalf("Failed to create sync token: %v", err)
	}

	fmt.Println("\n✅ Sync token created successfully!")
	fmt.Println("\n⚠️  IMPORTANT: Save this token securely - it will NOT be shown again!")
	fmt.Printf("\nToken: %s\n", response.Token)
	fmt.Printf("\nToken ID: %s\n", response.TokenInfo.ID)
	fmt.Printf("Tenant: %s (%s)\n", selectedTenant.Email, selectedTenant.FullName)
	fmt.Printf("Permissions: Read=%v, Write=%v, Delete=%v\n", canRead, canWrite, canDelete)
	fmt.Printf("Expires: %s\n", response.TokenInfo.ExpiresAt.Format("2006-01-02 15:04:05"))
}
