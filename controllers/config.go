package controllers

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// configEnv is the name of environment variable with config file path
	configEnv = "CONFIG_FILE"
)

// GroupVersionResourceConditions is a Kubernetes resource type along with a list of conditions
type GroupVersionResourceConditions struct {
	schema.GroupVersionResource
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition is the condition within resource status
type Condition struct {
	// Name of the condition
	Name string `json:"name"`
	// Status of the condition
	Status string `json:"status"`
}

// Config defines the configuration of the controllers
type Config struct {
	// ResourceTypes map from shortnames of Kubernetes API resources to their GVRs with conditions
	ResourceTypes map[string]GroupVersionResourceConditions `json:"resourceTypes,omitempty"`
	// DefaultResync period for controller watch functions
	DefaultResync string `json:"defaultResync,omitempty"`
	// ClusterScoped is true if Iter8 controller is cluster-scoped
	ClusterScoped bool `json:"clusterScoped,omitempty"`
	// Persist is true if Iter8 controller should have a persistent volume
	Persist bool `json:"persist,omitempty"`
	// Storage is the minimum amount of space that can be requested for the persistent volume
	Storage string `json:"storage,omitempty"`
	// StorageClassName is the provisioner for the persistent volume
	StorageClassName string `json:"storageClassName,omitempty"`
}

// readConfig reads configuration information from file
func readConfig() (*Config, error) {
	// read controller config
	configFile, ok := os.LookupEnv(configEnv)
	if !ok {
		e := errors.New("cannot lookup config env variable: " + configEnv)
		log.Logger.Error(e)
		return nil, e
	}

	filePath := filepath.Clean(configFile)
	dat, err := os.ReadFile(filePath)

	if err != nil {
		e := errors.New("cannot read config file: " + configFile)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	conf := Config{}
	err = yaml.Unmarshal(dat, &conf)
	if err != nil {
		e := errors.New("cannot unmarshal YAML config file: " + configFile)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	return &conf, nil
}

// validate the config
// no-op for now
func (c *Config) validate() error {
	return nil
}
