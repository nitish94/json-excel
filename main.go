package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/natefinch/lumberjack"
	"json-excel/pkg/utils"
	"json-excel/pkg/validation"
)

const DataFile = "data/data_demo.json"

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

	// Create data directory
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

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

	// Serve advanced.html for /advanced
	http.HandleFunc("/advanced", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/advanced.html")
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // uses default mux
	}

	// Channel to listen for interrupt signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server started on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for interrupt signal
	<-done
	log.Println("Server is shutting down...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// cleanupOldFiles runs every hour and deletes data/data_*.json files older than 24 hours
func cleanupOldFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			files, err := filepath.Glob("data/data_*.json")
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
