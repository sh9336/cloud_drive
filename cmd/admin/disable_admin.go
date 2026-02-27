// cmd/admin/disable_admin.go
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	printHeader("Disable Super Admin Account")

	// Connect to database
	db := connectDB()
	defer db.Close()

	// Get admin email
	email := getAdminEmail()

	// Fetch admin details
	admin := getAdminDetails(db, email)
	if admin == nil {
		log.Fatal("❌ Admin not found")
	}

	// Check if already disabled
	if admin["status"] == "disabled" {
		log.Fatal("❌ Admin is already disabled")
	}

	// Display admin info
	fmt.Printf("\nAdmin found:\n")
	fmt.Printf("- Email: %s\n", admin["email"])
	fmt.Printf("- Status: %s\n", admin["status"])
	if admin["last_login"] != "" {
		fmt.Printf("- Last login: %s\n", admin["last_login"])
	}

	// Get reason for disabling
	reason := getDisableReason()

	// Confirm action
	fmt.Println("\nThis will:")
	fmt.Println("  ✓ Set status to 'disabled'")
	fmt.Println("  ✓ Invalidate all existing sessions")
	fmt.Println("  ✓ Prevent future logins")
	fmt.Println("  ✓ Preserve all data and audit logs")
	fmt.Println("  ✓ Can be re-enabled later")

	if !confirmAction("DISABLE") {
		fmt.Println("\n❌ Disable operation cancelled")
		os.Exit(0)
	}

	// Get current user for disabled_by
	currentUser := getSSHUser()

	// Invalidate sessions
	fmt.Println("\n⏳ Invalidating sessions...")
	_, err := db.Exec(`DELETE FROM sessions WHERE admin_id = $1`, admin["id"])
	if err != nil {
		log.Printf("⚠️  Warning: Failed to invalidate sessions: %v", err)
	}

	// Update admin status
	fmt.Println("⏳ Updating admin status...")
	_, err = db.Exec(`
		UPDATE admins 
		SET status = 'disabled',
		    disabled_at = $1,
		    disabled_by = $2,
		    disable_reason = $3,
		    updated_at = $1
		WHERE id = $4
	`, time.Now(), currentUser, reason, admin["id"])
	if err != nil {
		log.Fatal("❌ Failed to disable admin:", err)
	}

	// Create audit log
	fmt.Println("⏳ Creating audit log...")
	logAudit(db, "admin_disabled", email, admin["id"].(string), reason)

	// Success message
	printSuccess()
	fmt.Printf("\nDisabled:\n")
	fmt.Printf("- Admin ID: %s\n", admin["id"])
	fmt.Printf("- Email: %s\n", admin["email"])
	fmt.Printf("- Status: disabled\n")

	fmt.Printf("\nAudit Trail:\n")
	fmt.Printf("- Action: admin_disabled\n")
	fmt.Printf("- Actor: system (SSH user: %s)\n", currentUser)
	fmt.Printf("- Reason: %s\n", reason)
	fmt.Printf("- Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))

	fmt.Println("\nTo re-enable:")
	fmt.Println("  go run cmd/admin/enable_admin.go")
}

func printHeader(title string) {
	fmt.Println("\n╔═══════════════════════════════════════════╗")
	fmt.Printf("║  %-38s  ║\n", title)
	fmt.Println("╚═══════════════════════════════════════════╝")
}

func printSuccess() {
	fmt.Println("\n╔═══════════════════════════════════════════╗")
	fmt.Println("║  ✅ Admin Disabled Successfully           ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
}

func getAdminEmail() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter admin email to disable: ")
	email, _ := reader.ReadString('\n')
	return strings.TrimSpace(email)
}

func getDisableReason() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nReason for disabling: ")
	reason, _ := reader.ReadString('\n')
	return strings.TrimSpace(reason)
}

func getAdminDetails(db *sql.DB, email string) map[string]interface{} {
	var id, adminEmail, status string
	var lastLogin sql.NullTime

	err := db.QueryRow(`
		SELECT id, email, status, last_login 
		FROM admins 
		WHERE email = $1
	`, email).Scan(&id, &adminEmail, &status, &lastLogin)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Fatal("❌ Database error:", err)
	}

	admin := map[string]interface{}{
		"id":     id,
		"email":  adminEmail,
		"status": status,
	}

	if lastLogin.Valid {
		admin["last_login"] = lastLogin.Time.Format("2006-01-02 15:04:05 MST")
	} else {
		admin["last_login"] = ""
	}

	return admin
}

func confirmAction(confirmWord string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nType '%s' to confirm: ", confirmWord)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input) == confirmWord
}

func connectDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable not set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}
	return db
}

func getSSHUser() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("SUDO_USER")
	}
	if user == "" {
		user = "unknown"
	}
	return user
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func logAudit(db *sql.DB, action, targetEmail, resourceID, reason string) {
	metadata := fmt.Sprintf(`{
		"ssh_user": "%s",
		"hostname": "%s",
		"reason": "%s"
	}`, getSSHUser(), getHostname(), reason)

	_, err := db.Exec(`
		INSERT INTO audit_logs (
			id, actor_type, actor_id, action, resource_type, 
			resource_id, metadata, created_at
		)
		VALUES (
			$1, 'system', 
			(SELECT id FROM admins WHERE email = $2 LIMIT 1),
			$3, 'admin', $4, $5::jsonb, $6
		)
	`, uuid.New().String(), targetEmail, action, resourceID, metadata, time.Now())

	if err != nil {
		log.Printf("⚠️  Warning: Failed to create audit log: %v", err)
	}
}
