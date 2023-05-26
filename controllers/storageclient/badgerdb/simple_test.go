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

	assert.NotNil(t, client)
	assert.NotNil(t, client.db)

	err = client.db.Close()
	assert.NoError(t, err)
}
