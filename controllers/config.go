package controllers

import (
	"errors"
	"os"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// configEnv is the name of environment variable with config file path
	configEnv = "CONFIG_FILE"
)

type GroupVersionKindResource struct {
	Group      string      `json:"group,omitempty"`
	Version    string      `json:"version,omitempty"`
	Resource   string      `json:"resource,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type Condition struct {
	Name   string `json:"name"`
	Status string `json:"name"`
}

// placeholders
type Config struct {
	AppNamespace string `json:"appNamespace,omitempty"`
	// map from shortnames of Kubernetes API resources to their GVRs
	KnownGVKRs map[string]GroupVersionKindResource `json:"knowngvkrs,omitempty"`
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

func (c *Config) mapGVKToGVR(gvk schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	for _, gvkr := range c.KnownGVKRs {
		if gvkr.Group == gvk.Group && gvkr.Version == gvk.Version && gvkr.Kind == gvk.Kind {
			return &schema.GroupVersionResource{
				Group:    gvkr.Group,
				Version:  gvkr.Version,
				Resource: gvkr.Resource,
			}, nil
		}
	}
	err := errors.New("unable to map gvk to gvr: " + gvk.String())
	log.Logger.Error(err)
	return nil, err
}

func (g *GroupVersionKindResource) matches(u *unstructured.Unstructured) bool {
	return g.Kind == u.GetKind() &&
		u.GetAPIVersion() ==
			(schema.GroupVersion{
				Group:   g.Group,
				Version: g.Version,
			}).String()
}
