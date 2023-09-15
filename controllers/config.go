package controllers

import (
	util "github.com/iter8-tools/iter8/base"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// configEnv is the name of environment variable with config file path
	configEnv = "CONFIG_FILE"
)

// GroupVersionResourceConditions is a Kubernetes resource type along with a list of conditions
type GroupVersionResourceConditions struct {
	schema.GroupVersionResource
	Conditions []string `json:"conditions,omitempty"`
}

// Config defines the configuration of the controllers
type Config struct {
	// ResourceTypes map from shortnames of Kubernetes API resources to their GVRs with conditions
	ResourceTypes map[string]GroupVersionResourceConditions `json:"resourceTypes,omitempty"`
	// DefaultResync period for controller watch functions
	DefaultResync string `json:"defaultResync,omitempty"`
	// ClusterScoped is true if Iter8 controller is cluster-scoped
	ClusterScoped bool `json:"clusterScoped,omitempty"`
}

// readConfig reads configuration information from file
func readConfig() (*Config, error) {
	conf := &Config{}
	err := util.ReadConfig(configEnv, conf, func() {})
	return conf, err
}

// validate the config
// no-op for now
func (c *Config) validate() error {
	return nil
}
