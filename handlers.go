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
var undoData sync.Map

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

// normalizeJSONData normalizes the JSON data to have consistent structure
func normalizeJSONData(data interface{}) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("data must be an array of objects")
	}
	if len(arr) == 0 {
		return arr, nil
	}

	// Collect all unique keys and their types
	keyTypes := make(map[string]string) // "primitive" or "nested"
	allKeys := make(map[string]bool)
	nestedSubKeys := make(map[string][]string)

	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("all items must be objects")
		}
		for key, val := range obj {
			allKeys[key] = true
			if arrVal, isArr := val.([]interface{}); isArr {
				keyTypes[key] = "nested"
				if len(arrVal) > 0 {
					if subObj, ok := arrVal[0].(map[string]interface{}); ok {
						subKeys := make([]string, 0)
						for subKey := range subObj {
							subKeys = append(subKeys, subKey)
						}
						nestedSubKeys[key] = subKeys
					}
				}
			} else if _, isObj := val.(map[string]interface{}); isObj {
				keyTypes[key] = "nested"
			} else if keyTypes[key] != "nested" {
				keyTypes[key] = "primitive"
			}
		}
	}

	// Normalize each object
	for i, item := range arr {
		obj := item.(map[string]interface{})
		for key := range allKeys {
			if _, exists := obj[key]; !exists {
				if keyTypes[key] == "nested" {
					obj[key] = []interface{}{}
				} else {
					obj[key] = nil
				}
			} else {
				// Ensure type consistency
				val := obj[key]
				if keyTypes[key] == "nested" {
					if arrVal, isArr := val.([]interface{}); isArr {
						// Ensure sub-objects have consistent keys
						subKeys := nestedSubKeys[key]
						for j, subItem := range arrVal {
							if subObj, ok := subItem.(map[string]interface{}); ok {
								for _, subKey := range subKeys {
									if _, exists := subObj[subKey]; !exists {
										subObj[subKey] = nil
									}
								}
							} else {
								// If not object, perhaps make it object with subKeys
								newObj := make(map[string]interface{})
								for _, subKey := range subKeys {
									newObj[subKey] = nil
								}
								arrVal[j] = newObj
							}
						}
					} else if objVal, isObj := val.(map[string]interface{}); isObj {
						// Single object, ensure keys
						subKeys := nestedSubKeys[key]
						for _, subKey := range subKeys {
							if _, exists := objVal[subKey]; !exists {
								objVal[subKey] = nil
							}
						}
					} else {
						// Convert primitive to empty array
						obj[key] = []interface{}{}
					}
				}
			}
		}
		arr[i] = obj
	}

	return arr, nil
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
		// Save current data for undo
		currentData, _ := utils.ReadJSONFile(dataFile)
		undoData.Store(id, currentData)

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

	// Normalize the data
	normalizedData, err := normalizeJSONData(newData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Normalization Error: %s", err), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateJSONStructure(normalizedData); err != nil {
		http.Error(w, fmt.Sprintf("Validation Error: %s", err), http.StatusBadRequest)
		return
	}

	// Generate Unique ID
	id := utils.GenerateID()
	targetFile := getDataFilePath(id)
	mu := getMutex(id)

	// Save
	mu.Lock()
	err = utils.WriteJSONFile(targetFile, normalizedData)
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

func undoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
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

	mu.Lock()
	defer mu.Unlock()

	// Get undo data
	undoVal, ok := undoData.Load(id)
	if !ok {
		http.Error(w, "No undo data available", http.StatusBadRequest)
		return
	}
	undoData.Delete(id) // One time use

	// Restore
	if err := utils.WriteJSONFile(dataFile, undoVal); err != nil {
		http.Error(w, "Error restoring file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
