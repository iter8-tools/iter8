package core

import (
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

// timePointer takes time.Time object as input, creates a new variable with the input value, and returns a pointer to the variable
func timePointer(t time.Time) *time.Time {
	return &t
}

// testingPatternPointer takes a TestingPattern value as input, creates a new variable with the input value, and returns a pointer to the variable
func testingPatternPointer(t TestingPatternType) *TestingPatternType {
	return &t
}
