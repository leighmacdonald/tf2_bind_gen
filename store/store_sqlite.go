package store

import (
	"bind_generator/model"
	"database/sql"
	"github.com/leighmacdonald/steamid"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type SQLiteDataStore struct {
	conn *sql.DB
}

func (s *SQLiteDataStore) setup() error {
	tablesQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name=?", (table_name,)`
	schema := `
		CREATE TABLE kills
            (
                kill_id INTEGER PRIMARY KEY AUTOINCREMENT,
                steam_id TEXT NOT NULL,
                weapon TEXT NOT NULL,
                is_crit BOOLEAN,
                created_on TIMESTAMP
            );
            CREATE TABLE messages
            (
                msg_id INTEGER PRIMARY KEY AUTOINCREMENT,
                msg TEXT NOT NULL,
                created_on TIMESTAMP
            );
	`
	if _, err := s.conn.Exec(schema); err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDataStore) Open(dsn string) error {
	c, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	s.conn = c
	if err := s.setup(); err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDataStore) Close() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			log.Errorf("Failed to close database")
		}
	}
	return nil
}

func (s *SQLiteDataStore) AddMessage(m string) error {
	return nil
}

func (s *SQLiteDataStore) TotalKills(sid steamid.SID64s) int {
	return 0
}
func (s *SQLiteDataStore) AddKillEvent(e *model.LogEvent) error {
	return nil
}
func (s *SQLiteDataStore) MigrateSID(from string, to string) error {
	return nil
}
