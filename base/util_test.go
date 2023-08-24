package base

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type config struct {
	Property *int `json:"property,omitempty"`
}

func TestReadConfigDefaultProperty(t *testing.T) {
	configEnvironnentVariable := "CONFIG"
	defaultPropertyValue := 8888

	file, err := os.CreateTemp("/tmp", "test")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	defer func() {
		err := os.Remove(file.Name())
		assert.NoError(t, err)
	}()

	err = os.Setenv(configEnvironnentVariable, file.Name())
	assert.NoError(t, err)
	conf := &config{}
	err = ReadConfig(configEnvironnentVariable, conf, func() {
		if nil == conf.Property {
			conf.Property = IntPointer(defaultPropertyValue)
		}
	})
	assert.NoError(t, err)

	assert.Equal(t, defaultPropertyValue, *conf.Property)
}

func TestReadConfigNoEnvVar(t *testing.T) {
	configEnvironnentVariable := "CONFIG"
	defaultPropertyValue := 8888

	// don't set environment variable
	conf := &config{}
	err := ReadConfig(configEnvironnentVariable, conf, func() {
		if nil == conf.Property {
			conf.Property = IntPointer(defaultPropertyValue)
		}
	})
	assert.Error(t, err)
}

func TestReadConfigNoFile(t *testing.T) {
	configEnvironnentVariable := "CONFIG"
	defaultPropertyValue := 8888

	err := os.Setenv(configEnvironnentVariable, "/tmp/noexistant")
	assert.NoError(t, err)
	conf := &config{}
	err = ReadConfig(configEnvironnentVariable, conf, func() {
		if nil == conf.Property {
			conf.Property = IntPointer(defaultPropertyValue)
		}
	})
	assert.Error(t, err)
}

func TestSplitApplication(t *testing.T) {
	ns, n := SplitApplication("namespace/name")
	assert.Equal(t, "namespace", ns)
	assert.Equal(t, "name", n)

	ns, n = SplitApplication("namespace/name/ignored")
	assert.Equal(t, "namespace", ns)
	assert.Equal(t, "name", n)

	ns, n = SplitApplication("name")
	assert.Equal(t, "default", ns)
	assert.Equal(t, "name", n)
}

type testType struct {
	S      string
	I      int
	Nested struct {
		S string
		I int
	}
}

func TestToYAML(t *testing.T) {
	obj := testType{
		S: "hello world",
		I: 3,
		Nested: struct {
			S string
			I int
		}{
			S: "nested",
		},
	}

	objString := ToYAML(obj)
	assert.Equal(t, `I: 3
Nested:
  I: 0
  S: nested
S: hello world`, string(objString))
}
