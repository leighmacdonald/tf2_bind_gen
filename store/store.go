package store

import (
	"bind_generator/model"
	"bind_generator/steam"
)

var store DataStoreI

type DataStoreI interface {
	Open(dsn string) error
	Close() error
	TotalKills(s steam.SID64s) int
	AddKillEvent(e model.LogEvent) error
	AddMessage(m string) error
	MigrateSID(from string, to string) error
}

func init() {
	store = &SQLiteDataStore{}
}
