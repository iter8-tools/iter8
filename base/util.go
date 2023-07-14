package base

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
)

// MajorMinor is the minor version of Iter8
// set this manually whenever the major or minor version changes
var MajorMinor = "v0.14"

// Version is the semantic version of Iter8 (with the `v` prefix)
// Version is intended to be set using LDFLAGS at build time
var Version = "v0.14.0"

const (
	toYAMLString = "toYaml"
)

// int64Pointer takes an int64 as input, creates a new variable with the input value, and returns a pointer to the variable
func int64Pointer(i int64) *int64 {
	return &i
}

// IntPointer takes an int as input, creates a new variable with the input value, and returns a pointer to the variable
func IntPointer(i int) *int {
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

// getTextTemplateFromURL gets template from URL
func getTextTemplateFromURL(providerURL string) (*template.Template, error) {
	// fetch b from url
	// #nosec
	resp, err := http.Get(providerURL)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	// read responseBody
	// get the doubly templated metrics spec
	defer func() {
		_ = resp.Body.Close()
	}()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tpl, err := CreateTemplate(string(responseBody))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	return tpl, nil
}

// CreateTemplate creates a template from a string
func CreateTemplate(tplString string) (*template.Template, error) {
	return template.New("provider template").Funcs(FuncMapWithToYAML()).Parse(string(tplString))
}

// FuncMapWithToYAML return sprig text function map with a toYaml function
func FuncMapWithToYAML() template.FuncMap {
	f := sprig.TxtFuncMap()
	f[toYAMLString] = ToYAML

	return f
}

// ToYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func ToYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// ReadConfig reads yaml formatted configuration information into conf
// from the file specified by environment variable configEnv
// The function setDefaults is called to set any default values if desired
func ReadConfig(configEnv string, conf interface{}, setDefaults func()) error {
	// identify location of config file from environment variable
	configFile, ok := os.LookupEnv(configEnv)
	if !ok {
		e := fmt.Errorf("environment variable %s not set", configEnv)
		log.Logger.Error(e)
		return e
	}

	// read the config file
	filePath := filepath.Clean(configFile)
	dat, err := os.ReadFile(filePath)

	if err != nil {
		e := errors.New("cannot read config file: " + configFile)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// convert to yaml
	err = yaml.Unmarshal(dat, &conf)
	if err != nil {
		e := errors.New("cannot unmarshal YAML config file: " + configFile)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// set any defaults for unset values
	setDefaults()
	return nil
}

// SplitApplication is a utility function that returns the namespace and name from a key of the form "namespace/name"
func SplitApplication(applicationKey string) (namespace string, name string) {
	names := strings.Split(applicationKey, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
}
