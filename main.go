package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const DataFile = "data.json"

func main() {
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
		WriteJSONFile(DataFile, initialData)
	}

	// Static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Endpoints
	http.HandleFunc("/api/data", dataHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/download", downloadHandler)

	fmt.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		data, err := ReadJSONFile(DataFile)
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

		if err := WriteJSONFile(DataFile, newData); err != nil {
			http.Error(w, fmt.Sprintf("Validation Error: %s", err), http.StatusBadRequest)
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

	if err := ValidateJSONStructure(newData); err != nil {
		http.Error(w, fmt.Sprintf("Validation Error: %s", err), http.StatusBadRequest)
		return
	}

	// Save
	// Re-marshal to pretty print and standarize
	if err := WriteJSONFile(DataFile, newData); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "File uploaded and saved."})
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := ReadJSONFile(DataFile)
	if err != nil {
		http.Error(w, "Error reading data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=data.json")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}
