package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllRoutemaps(t *testing.T) {
	rm := DefaultRoutemaps{}
	assert.NotNil(t, rm.GetAllRoutemaps())
}
