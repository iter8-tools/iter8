package watcher

import (
	"github.com/iter8-tools/iter8/base/log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// InformerClient is type of client to watch resources
type InformerClient struct {
	DC dynamic.Interface
}

// NewClient returns a new InformerClient that can be used to watch resources
func NewClient(cfg *rest.Config) (*InformerClient, error) {
	client := &InformerClient{}

	var err error
	// Grab a dynamic interface that we can create informers from
	client.DC, err = dynamic.NewForConfig(cfg)
	if err != nil {
		log.Logger.WithError(err).Fatal("could not generate dynamic client for config")
		return nil, err
	}

	return client, nil
}
