package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"json-excel/pkg/utils"
	"json-excel/pkg/validation"
)

var fileMutexes sync.Map

// getDataFilePath returns the filename for a given ID
func getDataFilePath(id string) string {
	// Sanitize ID to prevent directory traversal
	// Simple alphanumeric check could be added here, but hex is safe
	return fmt.Sprintf("data_%s.json", id)
}

// getMutex returns the RWMutex for a given ID
func getMutex(id string) *sync.RWMutex {
	if mu, ok := fileMutexes.Load(id); ok {
		return mu.(*sync.RWMutex)
	}
	mu := &sync.RWMutex{}
	actual, loaded := fileMutexes.LoadOrStore(id, mu)
	if loaded {
		return actual.(*sync.RWMutex)
	}
	return mu
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from query string
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	dataFile := getDataFilePath(id)
	mu := getMutex(id)

	switch r.Method {
	case "GET":
		mu.RLock()
		data, err := utils.ReadJSONFile(dataFile)
		mu.RUnlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)

	case "POST":
		mu.Lock()
		var newData interface{}
		if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
			mu.Unlock()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validation
		if err := validation.ValidateJSONStructure(newData); err != nil {
			mu.Unlock()
			http.Error(w, fmt.Sprintf("Validation Error: %s", err), http.StatusBadRequest)
			return
		}

		if err := utils.WriteJSONFile(dataFile, newData); err != nil {
			mu.Unlock()
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	r.ParseMultipartForm(maxUploadSize)

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read content
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Parse and Validate
	var newData interface{}
	if err := json.Unmarshal(byteValue, &newData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateJSONStructure(newData); err != nil {
		http.Error(w, fmt.Sprintf("Validation Error: %s", err), http.StatusBadRequest)
		return
	}

	// Generate Unique ID
	id := utils.GenerateID()
	targetFile := getDataFilePath(id)
	mu := getMutex(id)

	// Save
	mu.Lock()
	err = utils.WriteJSONFile(targetFile, newData)
	mu.Unlock()
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "File uploaded and saved.",
		"id":      id,
	})
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' parameter", http.StatusBadRequest)
		return
	}

	dataFile := getDataFilePath(id)
	mu := getMutex(id)

	// Check if file exists
	mu.RLock()
	_, err := os.Stat(dataFile)
	if os.IsNotExist(err) {
		mu.RUnlock()
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	data, err := utils.ReadJSONFile(dataFile)
	mu.RUnlock()
	if err != nil {
		http.Error(w, "Error reading data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=data_%s.json", id))

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}
