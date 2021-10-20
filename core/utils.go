package core

// UInt32Pointer takes a uint32 as input, creates a new variable with the input value, and returns a pointer to the variable
func UInt32Pointer(u uint32) *uint32 {
	return &u
}

// Int32Pointer takes an int32 as input, creates a new variable with the input value, and returns a pointer to the variable
func Int32Pointer(i int32) *int32 {
	return &i
}

// Float32Pointer takes an float32 as input, creates a new variable with the input value, and returns a pointer to the variable
func Float32Pointer(f float32) *float32 {
	return &f
}

// Float64Pointer takes an float64 as input, creates a new variable with the input value, and returns a pointer to the variable
func Float64Pointer(f float64) *float64 {
	return &f
}

// StringPointer takes a string as input, creates a new variable with the input value, and returns a pointer to the variable
func StringPointer(s string) *string {
	return &s
}

// BoolPointer takes a bool as input, creates a new variable with the input value, and returns a pointer to the variable
func BoolPointer(b bool) *bool {
	return &b
}
