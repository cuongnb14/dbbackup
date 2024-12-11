package main

import (
	"dbbackup/backup"
	"dbbackup/config"
	"flag"
	"log"

	"github.com/robfig/cron/v3"
)

func main() {
	runNow := flag.Bool("now", false, "Run the backup process immediately")
	flag.Parse()

	cfg, err := config.ReadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	if *runNow {
		log.Println("Running backup immediately...")
		backup.PerformBackup(cfg)
		return
	}

	c := cron.New()
	_, err = c.AddFunc(cfg.Cron, func() {
		backup.PerformBackup(cfg)
	})
	if err != nil {
		log.Fatalf("Failed to schedule backup job: %v", err)
	}

	log.Println("Backup scheduler started with cron: " + cfg.Cron)
	c.Start()

	select {}
}
