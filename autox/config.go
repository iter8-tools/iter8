package autox

// config.go - reading of configurtion (list of resources/namespaces to watch)

import (
	"os"
	"path/filepath"

	"github.com/iter8-tools/iter8/base/log"

	"sigs.k8s.io/yaml"
)

// trigger specifies a Kubernetes resource object. When this Kubernetes resource object is created/updated/deleted, then the releaseGroupSpecs will be applied/deleted.
type trigger struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	Group string `json:"group,omitempty" yaml:"group,omitempty"`

	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	Resource string `json:"resource,omitempty" yaml:"resource,omitempty"`
}

// releaseSpec points to a particular Helm releaseSpec
type releaseSpec struct {
	// Name is the name of the Helm chart
	Name string `json:"name" yaml:"name"`

	// Values is the values of the Helm chart
	Values map[string]interface{} `json:"values" yaml:"values"`

	// Version is the version of the Helm chart
	Version string `json:"version" yaml:"version"`
}

// releaseGroupSpec is the configuration of all the Helm charts for a particular experiment group and their install trigger
type releaseGroupSpec struct {
	// Trigger defines when the ReleaseSpecs should be installed
	Trigger trigger `json:"trigger" yaml:"trigger"`

	// ReleaseSpecs is the set of Helm charts
	// the keys in ReleaseSpecs are identifiers for each releaseSpec (releaseSpecID)
	ReleaseSpecs map[string]releaseSpec `json:"releaseSpecs" yaml:"releaseSpecs"`
}

// config is the configuration for all the Helm charts and their triggers
type config struct {
	// Specs contains the releaseGroupSpecs, which contain the Helm charts and their triggers
	// the keys in Specs are identifiers for each releaseGroupSpec (releaseGroupSpecID)
	Specs map[string]releaseGroupSpec
}

// readConfig reads YAML autoX config file and converts to a config object
func readConfig(fn string) (c config) {
	// empty configuration
	c = config{}

	yfile, err := os.ReadFile(filepath.Clean(fn))
	if err != nil {
		log.Logger.Warnf("unable to read configuration file %s: %s", fn, err.Error())
		return c // empty configuration
	}

	log.Logger.Debugf("read configuration\n%s", string(yfile))

	err = yaml.Unmarshal(yfile, &c)
	if err != nil {
		log.Logger.Warnf("invalid configuration file %s: %s", fn, err.Error())
		return c // empty configuration
	}

	return c
}
