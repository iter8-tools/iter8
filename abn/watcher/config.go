package watcher

import (
	"io/ioutil"

	"github.com/iter8-tools/iter8/base/log"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"
)

// Config is ABn serivce configuration
type Config struct {
	// Namespaces is list of namespaces to watch
	Namespaces []string `yaml:"namespaces,omitemtpy"`
	// Resources is list of resoure types that should be watched in the namespaces
	Resources []schema.GroupVersionResource `yaml:"resources,omitemtpy"`
}

// ReadConfig reads yaml config file fn and converts to Config object
func ReadConfig(fn string) *Config {
	yfile, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Logger.Fatal(err)
	}

	var config Config
	err2 := yaml.Unmarshal(yfile, &config)
	if err2 != nil {
		log.Logger.Fatal(err2)
	}

	return &config
}
