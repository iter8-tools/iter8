package controllers

import (
	"os"
	"testing"

	util "github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
)

func TestGetPodName(t *testing.T) {
	var tests = []struct {
		a *string
		b string
		c bool
	}{
		{util.StringPointer("x-0"), "x-0", true},
		{util.StringPointer("x-y-0"), "x-y-0", true},
		{util.StringPointer("x-1"), "x-1", true},
		{util.StringPointer("x-y-1"), "x-y-1", true},
		{util.StringPointer("x"), "x", true},
		{util.StringPointer(""), "", false},
		{nil, "", false},
	}

	for _, e := range tests {
		if e.a == nil {
			_ = os.Unsetenv(podNameEnvVariable)
		} else {
			_ = os.Setenv(podNameEnvVariable, *e.a)
		}
		podName, ok := getPodName()
		assert.Equal(t, e.b, podName)
		assert.Equal(t, e.c, ok)
	}

}

func TestLeaderIsMe(t *testing.T) {
	var tests = []struct {
		a string
		b bool
		c bool
	}{
		{"x-0", true, false},
		{"x-y-0", true, false},
		{"x-1", false, false},
		{"x-y-1", false, false},
		{"x", false, false},
		{"", false, true},
	}

	for _, e := range tests {
		_ = os.Setenv(podNameEnvVariable, e.a)
		leaderStatus, err := leaderIsMe()
		assert.Equal(t, e.b, leaderStatus)
		if e.c {
			assert.Error(t, err)
		}
	}

}
