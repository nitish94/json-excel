package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Validation Constants
const (
	MaxKeysPerObject = 10
	MaxNestingLevel  = 1 // 0 is root, 1 is one level deep (Table inside Table)
)

// ValidateJSONStructure checks if the JSON adheres to the constraints:
// 1. Max 10 keys per object (KPIs)
// 2. Max 1 level of nesting of objects (arrays of objects are considered structure, not nesting level increase for this logic implies 'Table within Table')
//    Note: The user request says "only one nested josn is allowed" and "table inside a table".
//    We will interpret this as: Root > Object > Nested Object (allowed). Root > Object > Nested Object > Deep Nested Object (not allowed).
//    Actually, usually "Table inside Table" means Root is an Array of Objects. Each Object has keys. One key might be an Array of Objects.
//    Let's stick to: Root (Array or Object) -> Level 0.
//    If a value is a complex object/array, that's Level 1.
//    Anything deeper is invalid.
func ValidateJSONStructure(data interface{}) error {
	return validateRecursive(data, 0)
}

func validateRecursive(data interface{}, level int) error {
	if level > MaxNestingLevel {
		return fmt.Errorf("nesting level exceeds maximum allowed (%d)", MaxNestingLevel)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Check key count
		if len(v) > MaxKeysPerObject {
			return fmt.Errorf("object has %d keys, maximum allowed is %d", len(v), MaxKeysPerObject)
		}
		// Recurse
		for _, val := range v {
			if isComplex(val) {
				if err := validateRecursive(val, level+1); err != nil {
					return err
				}
			}
		}
	case []interface{}:
		// Arrays don't count as a nesting level themselves in terms of "Table inside Table" visualization usually,
		// but if it's an array of objects, those objects are at the current level.
		// However, to keep it strict:
		for _, val := range v {
			if isComplex(val) {
				// If we are iterating an array, the items in it are effectively at the *same* structural level as the array for depth purposes?
				// Or does an array wrapper count?
				// Let's assume strict depth: generic object/array = +1 depth.
				// But for a list of items (Excel rows), we usually don't count the root list.
				// Let's pass 'level' if it's the Root array, but 'level' if it's a nested array?
				// To be safe and simple: Root is level 0.
				// If `v` is the root array, we check items at Level 0.
				// If `v` is a nested array (inside an object), we check items at Level 1.
				if err := validateRecursive(val, level); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isComplex(v interface{}) bool {
	_, isMap := v.(map[string]interface{})
	_, isSlice := v.([]interface{})
	return isMap || isSlice
}

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
func WriteJSONFile(path string, data interface{}) error {
	// First validate
	if err := ValidateJSONStructure(data); err != nil {
		return err
	}

	file, _ := json.MarshalIndent(data, "", "  ")
	return ioutil.WriteFile(path, file, 0644)
}
