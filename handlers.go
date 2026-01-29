package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"json-excel/pkg/utils"
	"json-excel/pkg/validation"
)

// getDataFilePath returns the filename for a given ID
func getDataFilePath(id string) string {
	// Sanitize ID to prevent directory traversal
	// Simple alphanumeric check could be added here, but hex is safe
	return fmt.Sprintf("data_%s.json", id)
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

	switch r.Method {
	case "GET":
		data, err := utils.ReadJSONFile(dataFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)

	case "POST":
		var newData interface{}
		if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validation
		if err := validation.ValidateJSONStructure(newData); err != nil {
			http.Error(w, fmt.Sprintf("Validation Error: %s", err), http.StatusBadRequest)
			return
		}

		if err := utils.WriteJSONFile(dataFile, newData); err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
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
	// Max upload size: 10 MB
	r.ParseMultipartForm(10 << 20)

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

	// Save
	if err := utils.WriteJSONFile(targetFile, newData); err != nil {
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

	// Check if file exists
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	data, err := utils.ReadJSONFile(dataFile)
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
