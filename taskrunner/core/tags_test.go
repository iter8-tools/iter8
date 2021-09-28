package core

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInterpolate(t *testing.T) {
	tags := NewTags().
		With("name", "tester").
		With("revision", "revision1").
		With("container", "super-container")

	// success cases
	inputs := []string{
		// `hello {{index . "name"}}`,
		// "hello {{index .name}}",
		"hello {{.name}}",
		"hello {{.name}}{{.other}}",
	}
	for _, str := range inputs {
		interpolated, err := tags.Interpolate(&str)
		assert.NoError(t, err)
		assert.Equal(t, "hello tester", interpolated)
	}

	// failure cases
	inputs = []string{
		// bad braces,
		"hello {{{index .name}}",
		// missing '.'
		"hello {{name}}",
	}
	for _, str := range inputs {
		_, err := tags.Interpolate(&str)
		assert.Error(t, err)
	}

	// empty tags (success cases)
	str := "hello {{.name}}"
	tags = NewTags()
	interpolated, err := tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "hello ", interpolated)

	// secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"secretName": []byte("tester"),
		},
	}

	str = "hello {{.secret.secretName}}"
	tags = NewTags().WithSecret("secret", &secret)
	assert.Contains(t, tags.M, "secret")
	interpolated, err = tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "hello tester", interpolated)
}

func TestWithVersionRecommendedForPromotionDeprecated(t *testing.T) {
	var data []byte
	data, err := ioutil.ReadFile(filepath.Join("..", "testdata", "experiment1.yaml"))
	assert.NoError(t, err)
	exp := &v2alpha2.Experiment{}
	err = yaml.Unmarshal(data, exp)
	assert.NoError(t, err)
	tags := NewTags().WithRecommendedVersionForPromotionDeprecated(exp)
	assert.Equal(t, "revision1", tags.M["revision"])
}

func TestWithOutVersionRecommendedForPromotionDeprecated(t *testing.T) {
	var data []byte
	data, err := ioutil.ReadFile(filepath.Join("..", "testdata", "experiment1-norecommended.yaml"))
	assert.NoError(t, err)
	exp := &v2alpha2.Experiment{}
	err = yaml.Unmarshal(data, exp)
	assert.NoError(t, err)
	tags := NewTags().WithRecommendedVersionForPromotionDeprecated(exp)
	assert.NotContains(t, tags.M, "revision1")
	// assert.Equal(t, "revision1", tags.M["revision"])
}

func TestWithVersionDefaultRecommendedForPromotion(t *testing.T) {
	var data []byte
	data, err := ioutil.ReadFile(filepath.Join("..", "testdata", "experiment1.yaml"))
	assert.NoError(t, err)
	exp := &v2alpha2.Experiment{}
	err = yaml.Unmarshal(data, exp)
	assert.NoError(t, err)

	vi := []VersionInfo{
		{Variables: []v2alpha2.NamedValue{{Name: "foo", Value: "bar1"}}},
		{Variables: []v2alpha2.NamedValue{{Name: "foo", Value: "bar2"}}},
	}

	tags := NewTags().WithRecommendedVersionForPromotion(exp, vi)
	assert.Equal(t, "bar1", tags.M["foo"])
}

func TestWithVersionCanaryRecommendedForPromotion(t *testing.T) {
	var data []byte
	data, err := ioutil.ReadFile(filepath.Join("..", "testdata", "experiment1-canarywinner.yaml"))
	assert.NoError(t, err)
	exp := &v2alpha2.Experiment{}
	err = yaml.Unmarshal(data, exp)
	assert.NoError(t, err)

	vi := []VersionInfo{
		{Variables: []v2alpha2.NamedValue{{Name: "foo", Value: "bar1"}}},
		{Variables: []v2alpha2.NamedValue{{Name: "foo", Value: "bar2"}}},
	}

	tags := NewTags().WithRecommendedVersionForPromotion(exp, vi)
	assert.Equal(t, "bar2", tags.M["foo"])
}

func TestWithOutVersionRecommendedForPromotion(t *testing.T) {
	var data []byte
	data, err := ioutil.ReadFile(filepath.Join("..", "testdata", "experiment1-norecommended.yaml"))
	assert.NoError(t, err)
	exp := &v2alpha2.Experiment{}
	err = yaml.Unmarshal(data, exp)
	assert.NoError(t, err)

	vi := []VersionInfo{
		{Variables: []v2alpha2.NamedValue{{Name: "foo", Value: "bar1"}}},
		{Variables: []v2alpha2.NamedValue{{Name: "foo", Value: "bar2"}}},
	}

	tags := NewTags().WithRecommendedVersionForPromotion(exp, vi)
	assert.NotContains(t, tags.M, "foo")
	// assert.Equal(t, "bar1", tags.M["foo"])
}
