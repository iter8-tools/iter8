package watcher

// config.go - reading of configurtion (list of resources/namespaces to watch)

import (
	"io/ioutil"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"
)

// Config is A/B(/n) serivce configuration. The configuration identifies the resources to watch and in which namespaces.
type Config struct {
	// Namespaces is list of namespaces to watch
	Namespaces []string `yaml:"namespaces,omitemtpy"`
	// Resources is list of resoure types that should be watched in the namespaces
	Resources []schema.GroupVersionResource `yaml:"resources,omitemtpy"`
}

// ReadConfig reads yaml config file fn and converts to a Config object
func ReadConfig(fn string) (config Config) {
	// empty configuration
	config = Config{
		Namespaces: []string{},
		Resources:  []schema.GroupVersionResource{},
	}

	yfile, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Logger.Warnf("unable to read configuration file %s: %s", fn, err.Error())
		return config // empty configuration
	}

	log.Logger.Debugf("read configuration\n%s", string(yfile))

	err = yaml.Unmarshal(yfile, &config)
	if err != nil {
		log.Logger.Warnf("invalid configuration file %s: %s", fn, err.Error())
		return config // empty configuration
	}

	if len(config.Namespaces) == 0 {
		log.Logger.Warn("not watching any namespaces - configuration error?")
	}
	if len(config.Resources) == 0 {
		log.Logger.Warn("not watching any resources - configuration error?")
	}

	return config
}
