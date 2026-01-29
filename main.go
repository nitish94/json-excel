package main

import (
	"encoding/json"
	"fmt"
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
