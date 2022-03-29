package base

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/iter8-tools/iter8/base/log"
)

// MajorMinor is the minor version of Iter8
// set this manually whenever the major or minor version changes
var MajorMinor = "v0.9"

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

// StringPointer takes string as input, creates a new variable with the input value, and returns a pointer to the variable
func StringPointer(s string) *string {
	return &s
}

// BoolPointer takes bool as input, creates a new variable with the input value, and returns a pointer to the variable
func BoolPointer(b bool) *bool {
	return &b
}

// CompletePath is a helper function for converting file paths, specified relative to the caller of this function, into absolute ones.
// CompletePath is useful in tests and enables deriving the absolute path of experiment YAML files.
func CompletePath(prefix string, suffix string) string {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(filename), prefix, suffix)
}

// getPayloadBytes downloads payload from URL and returns a byte slice
func getPayloadBytes(url string) ([]byte, error) {
	var myClient = &http.Client{}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	myClient = &http.Client{Transport: tr, Timeout: 10 * time.Second}

	r, err := myClient.Get(url)
	if err != nil || r.StatusCode >= 400 {
		e := errors.New("error while fetching payload")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, err
}

// getFileFromURL downloads contents from URL into the given file
func getFileFromURL(url string, fileName string) error {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil || r.StatusCode >= 400 {
		return errors.New("error while fetching payload")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, body, 0644)
	return err
}
