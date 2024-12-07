package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	_ "github.com/lib/pq"
)

func zipFolder(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func main() {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	sslmode := os.Getenv("PG_SSLMODE")

	if host == "" || port == "" || user == "" || password == "" || sslmode == "" {
		log.Fatalf("Missing required environment variables. Please set PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, and PG_SSLMODE.")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=postgres",
		host, port, user, password, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")
	if err != nil {
		log.Fatalf("Failed to query databases: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			log.Fatalf("Failed to scan database name: %v", err)
		}
		databases = append(databases, dbname)
	}

	backupDir := "./backup"
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.Mkdir(backupDir, 0755); err != nil {
			log.Fatalf("Failed to create backup directory: %v", err)
		}
	}

	for _, dbname := range databases {
		backupFile := fmt.Sprintf("%s/%s.sql", backupDir, dbname)
		log.Printf("Backing up database: %s", dbname)

		cmd := exec.Command("pg_dump",
			"-h", host,
			"-p", port,
			"-U", user,
			"-F", "c", // Compressed format
			"-f", backupFile,
			dbname,
		)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Failed to backup database %s: %v, output: %s", dbname, err, string(output))
		} else {
			log.Printf("Backup completed for database: %s", dbname)
		}
	}
}
