package storage

import (
	"fmt"
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
	assert.NotEqual(t, uint64(0), availableBytes)
	assert.NotEqual(t, uint64(0), totalBytes)

	availableBytes, totalBytes, err = GetVolumeUsage("non/existent/path")
	assert.Error(t, err)
	assert.Equal(t, uint64(0), totalBytes)
	assert.Equal(t, uint64(0), availableBytes)
}

type testmetrickey struct {
	valid       bool
	application string
	signature   string
	metric      string
	user        string
	transaction string
}

func TestGetMetricKey(t *testing.T) {
	for _, s := range []testmetrickey{
		{valid: true, application: "application", signature: "signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: false, application: "invalid:application", signature: "signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: true, application: "application", signature: "signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: false, application: "application", signature: "invalid:signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: true, application: "application", signature: "signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: false, application: "application", signature: "signature", metric: "invalid:metric", user: "user", transaction: "transaction"},
		{valid: true, application: "application", signature: "signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: false, application: "application", signature: "signature", metric: "metric", user: "invalid:user", transaction: "transaction"},
		{valid: true, application: "application", signature: "signature", metric: "metric", user: "user", transaction: "transaction"},
		{valid: false, application: "application", signature: "signature", metric: "metric", user: "user", transaction: "invalid:transaction"},
	} {
		key, err := GetMetricKey(s.application, 0, s.signature, s.metric, s.user, s.transaction)
		if s.valid {
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%s%s::%s::%s", GetMetricKeyPrefix(s.application, 0, s.signature), s.metric, s.user, s.transaction), key)
		} else {
			assert.Error(t, err)
			assert.Equal(t, "", key)
		}
	}
}

func TestValidateKeyToken(t *testing.T) {
	err := validateKeyToken("hello")
	assert.NoError(t, err)

	err = validateKeyToken("::")
	assert.Error(t, err)

	err = validateKeyToken("hello::world")
	assert.Error(t, err)

	err = validateKeyToken("hello :: world")
	assert.Error(t, err)

	err = validateKeyToken("hello:world")
	assert.Error(t, err)

	err = validateKeyToken("hello : world")
	assert.Error(t, err)
}

func TestGetUserPrefix(t *testing.T) {
	assert.Equal(t, "kt-users::app::0::abc::", GetUserKeyPrefix("app", 0, "abc"))
}

func TestGetUserKey(t *testing.T) {
	assert.Equal(t, "kt-users::app::0::abc::user", GetUserKey("app", 0, "abc", "user"))
}

func TestGetExperimentResultKey(t *testing.T) {
	assert.Equal(t, "kt-result::ns::name", GetExperimentResultKey("ns", "name"))
}

func TestGetExperimentResult(t *testing.T) {
	_, err := GetExperimentResult(func() ([]byte, error) { return []byte{}, nil })
	assert.ErrorContains(t, err, "cannot unmarshal ExperimentResult")

	_, err = GetExperimentResult(func() ([]byte, error) { return []byte{}, fmt.Errorf("test error") })
	assert.ErrorContains(t, err, "test error")
}
