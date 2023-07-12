package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVolumeUsage(t *testing.T) {
	// GetVolumeUsage is based off of statfs which analyzes the volume, not the directory
	// Creating a temporary directory will not change anything
	path, err := os.Getwd()
	assert.NoError(t, err)

	availableBytes, totalBytes, err := GetVolumeUsage(path)
	assert.NoError(t, err)

	// The volume should have some available and total bytes
	assert.NotEqual(t, 0, availableBytes)
	assert.NotEqual(t, 0, totalBytes)
}
