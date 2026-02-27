// cmd/admin/admin_audit_logs.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repository"
)

func main() {
	days := flag.Int("days", 7, "Number of days to look back")
	flag.Parse()

	fmt.Printf("=== Admin Audit Logs (Last %d Days) ===\n\n", *days)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.GetDSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewAuditRepository(db.DB)
	ctx := context.Background()

	logs, err := repo.List(ctx, *days)
	if err != nil {
		log.Fatalf("Failed to fetch audit logs: %v", err)
	}

	if len(logs) == 0 {
		fmt.Println("No audit logs found.")
		return
	}

	for _, logEntry := range logs {
		fmt.Printf("Time: %s\n", logEntry.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Actor: %s", logEntry.ActorType)
		if logEntry.ActorEmail != nil {
			fmt.Printf(" (%s)", *logEntry.ActorEmail)
		}
		fmt.Printf("\nAction: %s\n", logEntry.Action)
		if logEntry.ResourceType != nil {
			fmt.Printf("Resource: %s", *logEntry.ResourceType)
			if logEntry.ResourceID != nil {
				fmt.Printf(" (%s)", logEntry.ResourceID.String())
			}
			fmt.Println()
		}
		if len(logEntry.Metadata) > 0 {
			var metadata map[string]interface{}
			_ = json.Unmarshal(logEntry.Metadata, &metadata)
			metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
			fmt.Printf("Metadata: %s\n", string(metadataJSON))
		}
		fmt.Printf("Status: %s\n", logEntry.Status)
		fmt.Println("---")
	}
}
