package autox

// config.go - reading of configurtion (list of resources/namespaces to watch)

import (
	"os"
	"path/filepath"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"
)

// resourceConfig is the configuration that identifies the resources to watch and in which namespaces.
type resourceConfig struct {
	// Namespaces is list of namespaces to watch
	Namespaces []string `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`

	// Resources is list of resoure types that should be watched in the namespaces
	Resources []schema.GroupVersionResource `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// chart points to a particular Helm chart
type chart struct {
	// Repo is the repo of the Helm chart
	Repo string `json:"repo" yaml:"repo"`

	// Name is the name of the Helm chart
	Name string `json:"name" yaml:"name"`

	// ValuesFileURL is the URL to the values file of the Helm chart
	ValuesFileURL string `json:"valuesFileURL" yaml:"valuesFileURL"`
}

// chartGroup is the configuration of all the Helm charts for a particular experiment group
// the key is an identifier for the chart
type chartGroup struct {
	Charts map[string]chart `json:"charts" yaml:"charts"`
}

// chartGroupConfig is the configuration for all the Helm charts
// the key is the experiment group name
type chartGroupConfig map[string]chartGroup

// readResourceConfig reads yaml config file and converts to a resourceConfig object
func readResourceConfig(fp string) (config resourceConfig) {
	// empty configuration
	config = resourceConfig{}

	filePath := filepath.Clean(fp)
	yfile, err := os.ReadFile(filePath)
	if err != nil {
		log.Logger.Warnf("unable to read configuration file %s: %s", fp, err.Error())
		return config // empty configuration
	}

	log.Logger.Debugf("read configuration\n%s", string(yfile))

	err = yaml.Unmarshal(yfile, &config)
	if err != nil {
		log.Logger.Warnf("invalid configuration file %s: %s", fp, err.Error())
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

// readChartGroupConfig reads yaml chart group config file and converts to a chartGroupConfig object
func readChartGroupConfig(fn string) (config chartGroupConfig) {
	// empty configuration
	config = chartGroupConfig{}

	yfile, err := os.ReadFile(filepath.Clean(fn))
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
