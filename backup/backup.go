package backup

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"dbbackup/config"

	_ "github.com/lib/pq"
)

func PerformBackup(cfg *config.Config) {
	start := time.Now()

	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	sslmode := os.Getenv("PG_SSLMODE")

	if host == "" || port == "" || user == "" || password == "" || sslmode == "" {
		log.Fatalf("Missing required environment variables. Please set PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, and PG_SSLMODE.")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=postgres", host, port, user, password, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Get databases to backup
	databases := getDatabases(db, cfg.ExcludeDatabases)
	backupDir := fmt.Sprintf("/backup/%s", time.Now().Format("20060102_150405"))
	os.MkdirAll(backupDir, 0755)

	for _, dbname := range databases {
		backupDatabase(dbname, backupDir, cfg.ExcludeTables[dbname], host, port, user, password)
	}

	zipFile := fmt.Sprintf("%s.zip", backupDir)
	err = ZipFolder(backupDir, zipFile)
	if err != nil {
		log.Fatalf("Failed to zip backup directory: %v", err)
	}

	os.RemoveAll(backupDir)
	CleanupOldBackups("/backup", cfg.Keep)

	log.Printf("Backup completed in %s\n", time.Since(start))
}

func getDatabases(db *sql.DB, excludeList []string) []string {
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")
	if err != nil {
		log.Fatalf("Failed to query databases: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbname string
		rows.Scan(&dbname)
		if !isExcluded(dbname, excludeList) {
			databases = append(databases, dbname)
		}
	}
	return databases
}

func isExcluded(name string, excludeList []string) bool {
	for _, exclude := range excludeList {
		if name == exclude {
			return true
		}
	}
	return false
}

func backupDatabase(dbname, backupDir string, excludeTables []string, host, port, user, password string) {
	backupFile := fmt.Sprintf("%s/%s.backup", backupDir, dbname)
	cmd := exec.Command("pg_dump", "-h", host, "-p", port, "-U", user, "-F", "c", "-f", backupFile, dbname)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

	log.Printf("Backing up database: %s", dbname)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Failed to backup database %s: %v, output: %s", dbname, err, string(output))
	} else {
		log.Printf("Backup completed for database: %s", dbname)
	}
}
