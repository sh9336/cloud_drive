// cmd/admin/sync_token_stats.go
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
	fmt.Println("=== Sync Token Statistics ===\n")

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

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	// Get token ID
	fmt.Print("Sync Token ID: ")
	tokenIDStr, _ := reader.ReadString('\n')
	tokenIDStr = strings.TrimSpace(tokenIDStr)

	tokenID, err := uuid.Parse(tokenIDStr)
	if err != nil {
		log.Fatal("Invalid token ID")
	}

	// Get stats
	stats, err := syncTokenService.GetSyncTokenStats(ctx, tokenID)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}

	fmt.Printf("\n=== %s ===\n\n", stats.TokenName)
	fmt.Printf("Token ID: %s\n", stats.TokenID)
	fmt.Printf("Created: %s\n", stats.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Expires: %s (%d days left)\n", stats.ExpiresAt.Format("2006-01-02"), stats.DaysUntilExpiry)
	fmt.Printf("Status: %s\n\n", map[bool]string{true: "Active", false: "Inactive"}[stats.IsActive])

	fmt.Println("=== Usage Statistics ===")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Bytes Uploaded: %s\n", formatBytes(stats.TotalBytesUploaded))
	fmt.Printf("Bytes Downloaded: %s\n", formatBytes(stats.TotalBytesDownloaded))

	if stats.LastUsedAt != nil {
		fmt.Printf("Last Used: %s\n", stats.LastUsedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Last Used: Never")
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
