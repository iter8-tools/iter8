package watcher

// config.go - reading of configuration (list of resources/namespaces to watch)

import (
	"os"
	"path/filepath"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"
)

// // serviceConfig is an A/B/n service configuration. It is a map of namespace to a map of application details
type serviceConfig map[string]apps

// apps is a map of application name to service configuration details
type apps map[string]appDetails

// appDetails are the service configuration details for an application: the resource types and the max number of candidates
type appDetails struct {
	MaxNumCandidates int            `yaml:"maxNumCandidates,omitempty"`
	Resources        []resourceInfo `yaml:"resources,omitempty"`
}

type resourceInfo struct {
	schema.GroupVersionResource
	Condition *string `yaml:"condition,omitempty"`
}

// readServiceConfig reads yaml config file fn and converts to a Config object
func readServiceConfig(fn string) (c serviceConfig) {
	log.Logger.Tracef("readConfig called with config file %s", fn)

	// empty configuration
	c = serviceConfig{}

	yfile, err := os.ReadFile(filepath.Clean(fn))
	if err != nil {
		log.Logger.Warnf("unable to read configuration file %s: %s", fn, err.Error())
		return c // empty configuration
	}

	log.Logger.Debugf("read configuration file as:\n%s", string(yfile))

	err = yaml.Unmarshal(yfile, &c)
	if err != nil {
		log.Logger.Warnf("invalid configuration file %s: %s", fn, err.Error())
		return c // empty configuration
	}

	log.Logger.Tracef("readConfig returning config\n---\n%v\n---", c)
	return c
}

// getApplicationConfig extracts the application specific configuration from the service configuration
func getApplicationConfig(namespace string, application string, c serviceConfig) *appDetails {
	apps, ok := c[namespace]
	if !ok {
		// namespace not found, error
		log.Logger.Errorf("unable to find application configuration for %s/%s", namespace, application)
		return nil
	}

	appConfig, ok := apps[application]
	if !ok {
		log.Logger.Warnf("unable to find application configuration for %s/%s", namespace, application)
		return nil
	}

	return &appConfig
}
