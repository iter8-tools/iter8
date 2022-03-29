package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayload(t *testing.T) {
	// fetch jpeg image
	b, err := getPayloadBytes("https://cdn.pixabay.com/photo/2021/09/08/17/58/poppy-6607526_1280.jpg")
	assert.NoError(t, err)
	assert.NotNil(t, b)

	// fetch proto
	b, err = getPayloadBytes("https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto")
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
