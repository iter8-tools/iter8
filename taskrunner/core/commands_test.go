package core

import (
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/stretchr/testify/assert"
)

func TestGetActionSpec(t *testing.T) {
	var nilExp *Experiment = nil
	_, err := nilExp.GetActionSpec("stay-calm")
	assert.Error(t, err)

	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment8.yaml")).Build()
	assert.NoError(t, err)
	_, err = exp.GetActionSpec("stay-calm")
	assert.Error(t, err)

	a, err := exp.GetActionSpec("start")
	assert.NoError(t, err)
	assert.NotEmpty(t, a)

	exp, err = (&Builder{}).FromFile(CompletePath("../", "testdata/experiment4.yaml")).Build()
	assert.NoError(t, err)
	_, err = exp.GetActionSpec("start")
	assert.Error(t, err)
}

func TestUpdateVariable(t *testing.T) {
	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment6.yaml")).Build()
	assert.NoError(t, err)

	var v *v2alpha2.VersionDetail = nil
	err = UpdateVariable(v, "revision", "revision2")
	assert.Error(t, err)

	v = &exp.Spec.VersionInfo.Baseline
	err = UpdateVariable(v, "revision", "revision3")
	assert.Nil(t, err)

	err = UpdateVariable(v, "container", "turingmachine")
	assert.Nil(t, err)
	assert.Equal(t, "turingmachine", v.Variables[1].Value)
}

func TestGetVersionDetail(t *testing.T) {
	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment6.yaml")).Build()
	assert.NoError(t, err)

	v, err := exp.GetVersionDetail("default")
	assert.NotNil(t, v)
	assert.NoError(t, err)

	v, err = exp.GetVersionDetail("canary")
	assert.NotNil(t, v)
	assert.NoError(t, err)

	v, err = exp.GetVersionDetail("vpcandidate")
	assert.Nil(t, v)
	assert.Error(t, err)

	exp = nil
	v, err = exp.GetVersionDetail("vpcandidate")
	assert.Nil(t, v)
	assert.Error(t, err)
}

func TestFindVariableVersionDetail(t *testing.T) {
	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment6.yaml")).Build()
	assert.NoError(t, err)

	var v *v2alpha2.VersionDetail = nil
	val, err := FindVariableInVersionDetail(v, "revision")
	assert.Empty(t, val)
	assert.Error(t, err)

	v, err = exp.GetVersionDetail("default")
	assert.NotNil(t, v)
	assert.NoError(t, err)

	val, err = FindVariableInVersionDetail(v, "revision")
	assert.NotEmpty(t, val)
	assert.NoError(t, err)

	val, err = FindVariableInVersionDetail(v, "container")
	assert.Empty(t, val)
	assert.Error(t, err)
}
