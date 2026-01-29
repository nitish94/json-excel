package validation

import (
	"fmt"
)

// Validation Variables
var (
	MaxKeysPerObject = 10
	MaxNestingLevel  = 1 // 0 is root, 1 is one level deep (Table inside Table)
)

// ValidateJSONStructure checks if the JSON adheres to the constraints:
// 1. Max 10 keys per object (KPIs)
// 2. Max 1 level of nesting of objects
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
		for _, val := range v {
			if isComplex(val) {
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
