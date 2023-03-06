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

type GVR struct {
	// includes group, version, resource, and also readiness check
}

// placeholders
type Config struct {
	AppNamespaces []string `json:"appNamespaces,omitempty"`
	// short name for gvr to actual GVR
	GVRs map[string]GVR `json:"gvrs,omitempty"`
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
