// cmd/admin/create_admin.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/repository"

	"github.com/google/uuid"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

func printHeader() {
	fmt.Println(colorCyan + colorBold + "╔════════════════════════════════════════╗" + colorReset)
	fmt.Println(colorCyan + colorBold + "║                                        ║" + colorReset)
	fmt.Println(colorCyan + colorBold + "║     " + colorWhite + "🔐 Create Super Admin Tool" + colorCyan + "         ║" + colorReset)
	fmt.Println(colorCyan + colorBold + "║                                        ║" + colorReset)
	fmt.Println(colorCyan + colorBold + "╚════════════════════════════════════════╝" + colorReset)
	fmt.Println()
}

func printSuccess(message string) {
	fmt.Println(colorGreen + "✅ " + message + colorReset)
}

func printError(message string) {
	fmt.Println(colorRed + "❌ " + message + colorReset)
}

func printInfo(message string) {
	fmt.Println(colorBlue + "ℹ️  " + message + colorReset)
}

func printWarning(message string) {
	fmt.Println(colorYellow + "⚠️  " + message + colorReset)
}

func printPrompt(prompt string) {
	fmt.Print(colorPurple + "→ " + colorWhite + prompt + colorReset)
}

func printDivider() {
	fmt.Println(colorCyan + "────────────────────────────────────────" + colorReset)
}

func printSummary(admin *models.Admin) {
	fmt.Println()
	printDivider()
	fmt.Println(colorGreen + colorBold + "🎉 Super Admin Created Successfully!" + colorReset)
	printDivider()
	fmt.Println()
	fmt.Printf("  %s%-15s%s %s\n", colorCyan, "ID:", colorReset, admin.ID)
	fmt.Printf("  %s%-15s%s %s\n", colorCyan, "Email:", colorReset, admin.Email)
	fmt.Printf("  %s%-15s%s %s\n", colorCyan, "Full Name:", colorReset, admin.FullName)
	fmt.Printf("  %s%-15s%s %s\n", colorCyan, "Status:", colorReset, colorGreen+"Active"+colorReset)
	fmt.Printf("  %s%-15s%s %s\n", colorCyan, "Created At:", colorReset, admin.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	printDivider()
	printInfo("You can now use these credentials to log in to the admin panel.")
	fmt.Println()
}

func main() {
	printHeader()

	printInfo("Initializing database connection...")
	cfg, err := config.Load()
	if err != nil {
		printError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.GetDSN(), cfg.Database.MaxConns, cfg.Database.MinConns)
	if err != nil {
		printError(fmt.Sprintf("Failed to connect to database: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	printSuccess("Connected to database")
	fmt.Println()
	printDivider()
	fmt.Println()

	repo := repository.NewAdminRepository(db.DB)
	auditRepo := repository.NewAuditRepository(db.DB)
	reader := bufio.NewReader(os.Stdin)

	// Get email
	printPrompt("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email == "" {
		printError("Email is required")
		os.Exit(1)
	}

	// Check if admin already exists
	ctx := context.Background()
	existing, _ := repo.GetByEmail(ctx, email)
	if existing != nil {
		printError(fmt.Sprintf("Admin with email '%s' already exists", email))
		os.Exit(1)
	}

	// Get full name
	printPrompt("Full Name: ")
	fullName, _ := reader.ReadString('\n')
	fullName = strings.TrimSpace(fullName)

	if fullName == "" {
		printError("Full name is required")
		os.Exit(1)
	}

	// Get password
	printPrompt("Password (min 8 chars): ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if err := auth.ValidatePasswordStrength(password); err != nil {
		printError(fmt.Sprintf("Invalid password: %v", err))
		os.Exit(1)
	}

	fmt.Println()
	printInfo("Creating admin account...")

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		printError(fmt.Sprintf("Failed to hash password: %v", err))
		os.Exit(1)
	}

	// Create admin
	admin := &models.Admin{
		ID:                uuid.New(),
		Email:             email,
		PasswordHash:      passwordHash,
		FullName:          fullName,
		PasswordChangedAt: time.Now(),
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := repo.Create(ctx, admin); err != nil {
		printError(fmt.Sprintf("Failed to create admin: %v", err))
		os.Exit(1)
	}

	// Create audit log
	_ = auditRepo.Create(ctx, models.CreateAuditLogRequest{
		ActorType:    "system",
		ActorEmail:   &email,
		Action:       "admin_created",
		ResourceType: stringPtr("admin"),
		ResourceID:   &admin.ID,
		Metadata: map[string]interface{}{
			"created_via": "cli",
			"full_name":   fullName,
		},
		Status: "success",
	})

	printSummary(admin)
}

func stringPtr(s string) *string {
	return &s
}
