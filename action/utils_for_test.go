package action

import (
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

// secretDataReactor sets the secret.Data field based on the values from secret.StringData
// Credit: this function is adapted from https://github.com/creydr/go-k8s-utils
func secretDataReactor(action ktesting.Action) (bool, runtime.Object, error) {
	secret, _ := action.(ktesting.CreateAction).GetObject().(*corev1.Secret)

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	for k, v := range secret.StringData {
		sEnc := base64.StdEncoding.EncodeToString([]byte(v))
		secret.Data[k] = []byte(sEnc)
	}

	return false, nil, nil
}
