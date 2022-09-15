package core

import (
	"encoding/json"

	abnapp "github.com/iter8-tools/iter8/abn/application"
)

// getApplicationDataInternal is detailed implementation of gRPC method GetApplicationData
func getApplicationDataInternal(application string) (string, error) {
	a, err := abnapp.Applications.Get(application)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
