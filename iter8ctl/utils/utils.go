// Package utils is intended to contain helper functions used by other iter8ctl packages.
package utils

import (
	"encoding/json"
	"path/filepath"
	"runtime"
)

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, testFilename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(testFilename), prefix, suffix)
}

// IsJSON checks if the given string is a valid JSON object (schemaless check)
func IsJSONObject(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
