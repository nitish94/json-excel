package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/natefinch/lumberjack"
	"json-excel/pkg/utils"
	"json-excel/pkg/validation"
)

const DataFile = "data_demo.json"

var maxUploadSize int64

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	// Setup logging with rotation
	log.SetOutput(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	})

	// Get port from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get max keys from env
	maxKeysStr := os.Getenv("MAX_KEYS")
	maxKeys := 20
	if maxKeysStr != "" {
		if parsed, err := strconv.Atoi(maxKeysStr); err == nil {
			maxKeys = parsed
		}
	}
	validation.MaxKeysPerObject = maxKeys

	// Get max upload size from env (in MB)
	maxUploadMBStr := os.Getenv("MAX_UPLOAD_SIZE_MB")
	maxUploadMB := 1
	if maxUploadMBStr != "" {
		if parsed, err := strconv.Atoi(maxUploadMBStr); err == nil {
			maxUploadMB = parsed
		}
	}
	maxUploadSize = int64(maxUploadMB) * 1024 * 1024

	// Start cleanup goroutine
	go cleanupOldFiles()

	// Ensure data file exists or is handled
	if _, err := os.Stat(DataFile); os.IsNotExist(err) {
		// Create a sample file if not exists
		initialData := []map[string]interface{}{
			{
				"id":   1,
				"name": "Project Alpha",
				"kpis": []map[string]interface{}{
					{"metric": "Revenue", "value": 100},
					{"metric": "Cost", "value": 50},
				},
				"owner": "Alice",
			},
			{
				"id":    2,
				"name":  "Project Beta",
				"owner": "Bob",
				"kpis": []map[string]interface{}{
					{"metric": "Revenue", "value": 200},
				},
			},
		}
		utils.WriteJSONFile(DataFile, initialData)
	}

	// Static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Endpoints
	http.HandleFunc("/api/data", dataHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/download", downloadHandler)
	http.HandleFunc("/api/undo", undoHandler)

	log.Printf("Server started on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// cleanupOldFiles runs every hour and deletes data_*.json files older than 24 hours
func cleanupOldFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			files, err := filepath.Glob("data_*.json")
			if err != nil {
				log.Printf("Error globbing files: %v", err)
				continue
			}
			now := time.Now()
			for _, file := range files {
				info, err := os.Stat(file)
				if err != nil {
					log.Printf("Error stating file %s: %v", file, err)
					continue
				}
				if now.Sub(info.ModTime()) > 24*time.Hour {
					if err := os.Remove(file); err != nil {
						log.Printf("Error removing old file %s: %v", file, err)
					} else {
						log.Printf("Removed old file: %s", file)
					}
				}
			}
			log.Printf("Cleanup completed, checked %d files", len(files))
		}
	}
}
