package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocsCmd(t *testing.T) {
	// set COMMAND_DOCS_DIR
	dir, err := ioutil.TempDir("", "iter8docs")
	assert.NoError(t, err)

	commandDocsDir = dir
	err = docsCmd.RunE(nil, nil)
	assert.NoError(t, err)

	defer os.RemoveAll(dir)
}
