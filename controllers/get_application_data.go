package controllers

import (
	"encoding/json"
	"fmt"
)

// getApplicationDataInternal is detailed implementation of gRPC method GetApplicationData
func getApplicationDataInternal(application string) (string, error) {

	namespace, name := splitApplicationKey(application)
	s := allRoutemaps.getRoutemapFromNamespaceName(namespace, name)
	if s == nil {
		return "", fmt.Errorf("routemap not found for application %s", application)
	}

	legacyApp := routemapToLegacyApplication(s)

	jsonBytes, err := json.Marshal(legacyApp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
