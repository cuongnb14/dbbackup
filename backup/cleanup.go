package backup

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func CleanupOldBackups(backupDir string, keepCount int) {
	files, err := filepath.Glob(filepath.Join(backupDir, "*"))
	if err != nil {
		log.Fatalf("Failed to list backup files: %v", err)
	}

	type fileInfo struct {
		path string
		time time.Time
	}
	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("Failed to stat file: %s, error: %v", file, err)
			continue
		}
		fileInfos = append(fileInfos, fileInfo{path: file, time: info.ModTime()})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].time.After(fileInfos[j].time)
	})

	for i, file := range fileInfos {
		if i >= keepCount {
			log.Printf("Deleting old backup: %s", file.path)
			os.Remove(file.path)
		}
	}
}
