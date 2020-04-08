package store

import (
	"bind_generator"
	"bind_generator/steam"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDataStore struct {
	conn sql.Conn
}

func (s *SQLiteDataStore) Open(dsn string) error {
	c, err := sql.Open("sqlite3", dsn)
	return nil
}

func (s *SQLiteDataStore) Close() error {
	return nil
}

func (s *SQLiteDataStore) AddMessage(m string) error {
	return nil
}

func (s *SQLiteDataStore) TotalKills(sid steam.SID64s) int {
	return 0
}
func (s *SQLiteDataStore) AddKillEvent(e main.logEvent) error {
	return nil
}
func (s *SQLiteDataStore) MigrateSID(from string, to string) error {
	return nil
}
