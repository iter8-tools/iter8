package controllers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPodName(t *testing.T) {
	var tests = []struct {
		a string
		b string
		c bool
	}{
		{"x-0", "x-0", true},
		{"x-y-0", "x-y-0", true},
		{"x-1", "x-1", true},
		{"x-y-1", "x-y-1", true},
		{"x", "x", true},
		{"", "", false},
	}

	for _, e := range tests {
		_ = os.Setenv(PodNameEnvVariable, e.a)
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
		_ = os.Setenv(PodNameEnvVariable, e.a)
		leaderStatus, err := leaderIsMe()
		assert.Equal(t, e.b, leaderStatus)
		if e.c {
			assert.Error(t, err)
		}
	}

}
