package report

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// copyFileToPwd copies the specified file to pwd
func copyFileToPwd(t *testing.T, filePath string) error {
	// get file
	srcFile, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return errors.New("could not open metrics file")
	}
	t.Cleanup(func() {
		_ = srcFile.Close()
	})

	// create copy of file in pwd
	destFile, err := os.Create(filepath.Base(filePath))
	if err != nil {
		return errors.New("could not create copy of metrics file in temp directory")
	}
	t.Cleanup(func() {
		_ = destFile.Close()
	})
	_, _ = io.Copy(destFile, srcFile)
	return nil
}
