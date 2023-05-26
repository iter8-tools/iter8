package badgerdb

import (
	"os"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	tempDirPath := t.TempDir()
	_ = os.Chdir(tempDirPath)

	client, err := GetClient(badger.DefaultOptions(tempDirPath))
	assert.NoError(t, err)
	defer client.db.Close()

	assert.NotNil(t, client)
}
