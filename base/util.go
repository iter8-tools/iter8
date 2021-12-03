package base

import (
	"path/filepath"
	"runtime"
	"time"
)

// int64Pointer takes an int64 as input, creates a new variable with the input value, and returns a pointer to the variable
func int64Pointer(i int64) *int64 {
	return &i
}

// intPointer takes an int as input, creates a new variable with the input value, and returns a pointer to the variable
func intPointer(i int) *int {
	return &i
}

// float32Pointer takes an float32 as input, creates a new variable with the input value, and returns a pointer to the variable
func float32Pointer(f float32) *float32 {
	return &f
}

// float64Pointer takes an float64 as input, creates a new variable with the input value, and returns a pointer to the variable
func float64Pointer(f float64) *float64 {
	return &f
}

// stringPointer takes string as input, creates a new variable with the input value, and returns a pointer to the variable
func stringPointer(s string) *string {
	return &s
}

// timePointer takes time as input, creates a new variable with the input value, and returns a pointer to the variable
func timePointer(t time.Time) *time.Time {
	return &t
}

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(filename), prefix, suffix)
}
