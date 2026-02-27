// cmd/admin/list_admins.go
package main

import (
	"context"
	"fmt"
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repository"
)

func main() {
	fmt.Println("=== List All Super Admins ===\n")

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
	ctx := context.Background()

	admins, err := repo.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list admins: %v", err)
	}

	if len(admins) == 0 {
		fmt.Println("No admins found.")
		return
	}

	fmt.Printf("Total Admins: %d\n\n", len(admins))
	fmt.Println("ID                                   | Email                    | Name              | Status")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	for _, admin := range admins {
		status := "Active"
		if !admin.IsActive {
			status = "Disabled"
		}
		fmt.Printf("%-36s | %-24s | %-17s | %s\n",
			admin.ID.String(),
			admin.Email,
			truncate(admin.FullName, 17),
			status,
		)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
