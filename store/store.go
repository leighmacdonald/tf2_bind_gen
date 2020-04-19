package store

import (
	"bind_generator/model"
	"github.com/leighmacdonald/steamid"
	"os"
)

var store DataStoreI

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

type DataStoreI interface {
	Open(dsn string) error
	Close() error
	TotalKills(s steamid.SID64s) int
	AddKillEvent(e *model.LogEvent) error
	AddMessage(m string) error
	MigrateSID(from string, to string) error
}

func init() {
	store = &SQLiteDataStore{}
}
