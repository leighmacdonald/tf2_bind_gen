package store

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSQLiteDataStore(t *testing.T) {
	testDbPath := "./test_db.sqlite"
	if exists(testDbPath) {
		assert.NoError(t, os.Remove(testDbPath))
	}
	ds := SQLiteDataStore{}
	assert.NoError(t, ds.Open(testDbPath))
	assert.NoError(t, ds.AddMessage("msg 1"))
	assert.NoError(t, ds.AddMessage("msg 2"))
}
