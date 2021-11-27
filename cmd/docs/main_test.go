package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// set COMMAND_DOCS_DIR
	dir, err := ioutil.TempDir("", "iter8docs")
	assert.NoError(t, err)

	os.Setenv(commandDocsDir, dir)
	main()

	os.Setenv(commandDocsDir, "that-strange-place-which-doesnt-exist")
	main()

	defer os.RemoveAll(dir)
	// call main
}
