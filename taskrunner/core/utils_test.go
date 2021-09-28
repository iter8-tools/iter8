package core

import (
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetJsonBytes(t *testing.T) {
	// valid
	_, err := GetPayloadBytes("https://httpbin.org/stream/1")
	assert.NoError(t, err)

	// invalid
	_, err = GetPayloadBytes("https://httpbin.org/undef")
	assert.Error(t, err)
}

func TestPointers(t *testing.T) {
	assert.Equal(t, int32(1), *Int32Pointer(1))
	assert.Equal(t, float32(0.1), *Float32Pointer(0.1))
	assert.Equal(t, float64(0.1), *Float64Pointer(0.1))
	assert.Equal(t, "hello", *StringPointer("hello"))
	assert.Equal(t, false, *BoolPointer(false))
	assert.Equal(t, GET, *HTTPMethodPointer(GET))
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel(logrus.InfoLevel)
	assert.Equal(t, logrus.InfoLevel, log.GetLevel())
}

func TestIter8LogPrecedence(t *testing.T) {
	exp := &Experiment{
		Experiment: v2alpha2.Experiment{
			ObjectMeta: v1.ObjectMeta{
				Name:      "hello",
				Namespace: "default",
			},
			Spec:   v2alpha2.ExperimentSpec{},
			Status: v2alpha2.ExperimentStatus{},
		},
	}
	p := GetIter8LogPrecedence(exp, "start")
	assert.Equal(t, 0, p)
}
