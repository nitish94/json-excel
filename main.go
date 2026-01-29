package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"json-excel/pkg/utils"
)

const DataFile = "data_demo.json"

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
		utils.WriteJSONFile(DataFile, initialData)
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
