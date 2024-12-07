package main

import (
	"archive/zip"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ExcludeDatabases []string            `yaml:"exclude_databases"`
	ExcludeTables    map[string][]string `yaml:"exclude_tables"`
	Keep             int                 `yaml:"keep"`
	Cron             string              `yaml:"cron"`
}

// Read YAML configuration file
func readConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

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

func cleanupOldBackups(backupDir string, keepCount int) error {
	log.Printf("Keep: %d version", keepCount)

	files, err := filepath.Glob(filepath.Join(backupDir, "*.zip"))
	if err != nil {
		return err
	}

	type fileInfo struct {
		path string
		time time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, fileInfo{path: file, time: info.ModTime()})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].time.After(fileInfos[j].time)
	})

	for i, file := range fileInfos {
		if i >= keepCount {
			log.Printf("Deleting old backup: %s", file.path)
			if err := os.Remove(file.path); err != nil {
				log.Printf("Failed to delete file: %s, error: %v", file.path, err)
			}
		}
	}

	return nil
}

func performBackup(config *Config) {
	// Load configuration from YAML

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

		// Skip excluded databases
		excluded := false
		for _, exclude := range config.ExcludeDatabases {
			if dbname == exclude {
				excluded = true
				break
			}
		}
		if !excluded {
			databases = append(databases, dbname)
		}
	}

	backupTime := time.Now().Format("20060102_150405")
	backupDir := fmt.Sprintf("/backup/%s", backupTime)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.Mkdir(backupDir, 0755); err != nil {
			log.Fatalf("Failed to create backup directory: %v", err)
		}
	}

	for _, dbname := range databases {
		backupFile := fmt.Sprintf("%s/%s.backup", backupDir, dbname)
		log.Printf("Backing up database: %s", dbname)

		// Prepare exclude tables option
		excludedTables, ok := config.ExcludeTables[dbname]
		excludeOptions := ""
		if ok {
			for _, table := range excludedTables {
				excludeOptions += fmt.Sprintf("--exclude-table=%s ", table)
			}
		}

		cmd := exec.Command("pg_dump",
			"-h", host,
			"-p", port,
			"-U", user,
			"-F", "c", // Compressed format
			"-f", backupFile,
		)
		if excludeOptions != "" {
			cmd.Args = append(cmd.Args, excludeOptions)
		}
		cmd.Args = append(cmd.Args, dbname)

		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
		log.Println(cmd.String())
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Failed to backup database %s: %v, output: %s", dbname, err, string(output))
		} else {
			log.Printf("Backup completed for database: %s", dbname)
		}
	}

	zipFile := fmt.Sprintf("/backup/%s.zip", backupTime)
	if err := zipFolder(backupDir, zipFile); err != nil {
		log.Fatalf("Failed to create zip file: %v", err)
	}

	if err := os.RemoveAll(backupDir); err != nil {
		log.Printf("Failed to remove temporary backup directory: %v", err)
	} else {
		log.Printf("Temporary backup directory removed: %s", backupDir)
	}

	log.Printf("Backup completed and compressed to: %s", zipFile)

	if err := cleanupOldBackups("/backup", config.Keep); err != nil {
		log.Fatalf("Failed to cleanup old backups: %v", err)
	}

}

func main() {
	runNow := flag.Bool("now", false, "Run the backup process immediately")
	flag.Parse()

	config, err := readConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	if *runNow {
		log.Println("Running backup immediately...")
		performBackup(config)
		return
	}

	c := cron.New()

	_, err = c.AddFunc(config.Cron, func() {
		performBackup(config)
	})
	if err != nil {
		log.Fatalf("Failed to schedule backup job: %v", err)
	}

	log.Println("Backup scheduler started. Backup will run daily at 2:00 AM.")
	c.Start()

	select {}
}
