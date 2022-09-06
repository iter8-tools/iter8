package autox

// config.go - reading of configurtion (list of resources/namespaces to watch)

import (
	"io/ioutil"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"
)

// resourceConfig is the configuration that identifies the resources to watch and in which namespaces.
type resourceConfig struct {
	// Namespaces is list of namespaces to watch
	Namespaces []string `yaml:"namespaces,omitempty"`

	// Resources is list of resoure types that should be watched in the namespaces
	Resources []schema.GroupVersionResource `yaml:"resources,omitempty"`
}

type chart struct {
	// Repo is the repo of the helm chart
	Repo string `yaml:"repo"`

	// Name is the name of the helm chart
	Name string `yaml:"name"`

	// ValuesFileURL is the URL to the values file of the helm chart
	ValuesFileURL string `yaml:"valuesFileURL"`
}

type chartGroup struct {
	Name string `yaml:"name"`

	Charts []chart `yaml:"charts"`
}

type chartGroupConfig []chartGroup

// readConfig reads yaml config file fn and converts to a Config object
func readConfig(fn string) (config resourceConfig) {
	// empty configuration
	config = resourceConfig{}

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

// readChartGroupConfig reads yaml config file fn and converts to a Config object
func readChartGroupConfig(fn string) (config chartGroupConfig) {
	// empty configuration
	config = chartGroupConfig{}

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

	// if len(config.Namespaces) == 0 {
	// 	log.Logger.Warn("not watching any namespaces - configuration error?")
	// }
	// if len(config.Resources) == 0 {
	// 	log.Logger.Warn("not watching any resources - configuration error?")
	// }

	return config
}
