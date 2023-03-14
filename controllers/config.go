package controllers

import (
	"errors"
	"os"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// configEnv is the name of environment variable with config file path
	configEnv = "CONFIG_FILE"
)

type GroupVersionResourceConditions struct {
	Group      string      `json:"group,omitempty"`
	Version    string      `json:"version,omitempty"`
	Resource   string      `json:"resource,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type Condition struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// placeholders
type Config struct {
	AppNamespace *string `json:"appNamespace,omitempty"`
	// map from shortnames of Kubernetes API resources to their GVRs
	KnownGVRs map[string]GroupVersionResourceConditions `json:"knownGVRs,omitempty"`
	// Default Resync period for controller watch functions
	DefaultResync string `json:"defaultResync,omitempty"`
}

func readConfig() (*Config, error) {
	// read controller config
	configFile, ok := os.LookupEnv(configEnv)
	if !ok {
		e := errors.New("cannot lookup config env variable: " + configEnv)
		log.Logger.Error(e)
		return nil, e
	}

	dat, err := os.ReadFile(configFile)
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

func (c *Config) validate() error {
	return nil
}
