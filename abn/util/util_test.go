package util_test

import (
	"testing"

	"github.com/iter8-tools/iter8/abn/util"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for _, tt := range []struct {
		name   string
		length int
	}{{"one", 1}, {"five", 5}, {"nine", 9}} {
		t.Run(tt.name, func(t *testing.T) {
			a := util.RandomString(tt.length)
			assert.Equal(t, tt.length, len(a), "expected length %d", tt.length)
		})
	}
}
