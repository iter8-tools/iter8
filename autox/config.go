package autox

// config.go - reading of configurtion (list of resources/namespaces to watch)

import (
	"os"
	"path/filepath"

	"github.com/iter8-tools/iter8/base/log"

	"sigs.k8s.io/yaml"
)

// Opts are the options used for launching autoX service
type Opts struct {
	// KubeClient enables Kubernetes and Helm interactions with the cluster
	*kubeClient
}

// trigger specifies when a chartGroup should be installed
type trigger struct {
	Group string `json:"group,omitempty" yaml:"group,omitempty"`

	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	Resource string `json:"resource,omitempty" yaml:"resource,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// // chart points to a particular Helm chart
// type chart struct {
// 	// Repo is the repo of the Helm chart
// 	Repo string `json:"repo" yaml:"repo"`

// 	// Name is the name of the Helm chart
// 	Name string `json:"name" yaml:"name"`

// 	// ValuesFileURL is the URL to the values file of the Helm chart
// 	ValuesFileURL string `json:"valuesFileURL" yaml:"valuesFileURL"`
// }

// chart points to a particular Helm chart
type chart struct {
	// RepoURL is the url of the Helm repo
	RepoURL string `json:"repoURL" yaml:"repoURL"`

	// Name is the name of the Helm chart
	Name string `json:"name" yaml:"name"`

	// Values is the values of the Helm chart
	Values map[string]interface{} `json:"values" yaml:"values"`

	// Version is the version of the Helm chart
	// TODO: add version constraint, example: "1.16.X"
	Version string `json:"version" yaml:"version"`
}

// chartGroup is the configuration of all the Helm charts for a particular experiment group and their install trigger
type chartGroup struct {
	// Trigger defines when the ReleaseSpecs should be installed
	Trigger trigger `json:"trigger" yaml:"trigger"`

	// ReleaseSpecs is the set of Helm charts
	// the keys in ReleaseSpecs are identifiers for each chart
	ReleaseSpecs map[string]chart `json:"releaseSpecs" yaml:"releaseSpecs"`
}

// // chartGroupConfig is the configuration for all the Helm charts and their triggers
type chartGroupConfig struct {
	// Namespaces are the namespaces that autoX cleans on start
	Namespaces []string `json:"namespaces" yaml:"namespaces"`

	// Specs contains the chartGroups, which contain the Helm charts and their triggers
	Specs map[string]chartGroup
}

// NewOpts returns an autox options object
func NewOpts(kc *kubeClient) *Opts {
	return &Opts{
		kubeClient: kc,
	}
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
