package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// ReadJSONFile reads and parses the JSON file
func ReadJSONFile(path string) (interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty list if file doesn't exist
			return []interface{}{}, nil
		}
		return nil, err
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)
	var result interface{}
	if len(byteValue) == 0 {
		return []interface{}{}, nil
	}

	err = json.Unmarshal(byteValue, &result)
	return result, err
}

// WriteJSONFile writes the JSON data to file
// Note: Does NOT perform validation. Validation should be done before calling this.
func WriteJSONFile(path string, data interface{}) error {
	file, _ := json.MarshalIndent(data, "", "  ")
	return ioutil.WriteFile(path, file, 0644)
}
