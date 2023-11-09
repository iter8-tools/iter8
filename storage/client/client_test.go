package client

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/stretchr/testify/assert"
)

func TestGetClientRedis(t *testing.T) {

	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	metricsConfig := `port: 8080
implementation: redis
redis:
  address: ` + server.Addr()

	mf, err := os.CreateTemp("", "metrics*.yaml")
	assert.NoError(t, err)

	err = os.Setenv(metricsConfigFileEnv, mf.Name())
	assert.NoError(t, err)

	_, err = mf.WriteString(metricsConfig)
	assert.NoError(t, err)

	client, err := GetClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetClientBadger(t *testing.T) {

	tempDirPath := os.TempDir()

	metricsConfig := `port: 8080
implementation: badgerdb
badgerdb:
  dir: ` + tempDirPath

	mf, err := os.CreateTemp("", "metrics*.yaml")
	assert.NoError(t, err)

	err = os.Setenv(metricsConfigFileEnv, mf.Name())
	assert.NoError(t, err)

	_, err = mf.WriteString(metricsConfig)
	assert.NoError(t, err)

	client, err := GetClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
