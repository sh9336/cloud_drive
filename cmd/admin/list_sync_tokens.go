// cmd/admin/list_sync_tokens.go
package main

import (
	"context"
	"fmt"
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repository"
	"backend/internal/service"
)

func main() {
	fmt.Println("=== List All Sync Tokens ===\n")

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

	syncTokenService := service.NewSyncTokenService(syncTokenRepo, tenantRepo, auditRepo)

	ctx := context.Background()
	tokens, err := syncTokenService.ListAllSyncTokens(ctx)
	if err != nil {
		log.Fatalf("Failed to list sync tokens: %v", err)
	}

	if len(tokens) == 0 {
		fmt.Println("No sync tokens found.")
		return
	}

	fmt.Printf("Total Sync Tokens: %d\n\n", len(tokens))
	fmt.Println("ID                                   | Name              | Tenant               | Status   | Expires")
	fmt.Println("-----------------------------------------------------------------------------------------------------------")

	for _, token := range tokens {
		status := "Active"
		if !token.IsActive {
			status = "Revoked"
		}

		fmt.Printf("%-36s | %-17s | %-20s | %-8s | %s\n",
			token.ID.String(),
			truncate(token.Name, 17),
			truncate(token.TenantEmail, 20),
			status,
			token.ExpiresAt.Format("2006-01-02"),
		)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
