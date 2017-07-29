package server

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func setupDatabaseSqlite() (*sql.DB, error) {
	sq := Config.Database.Sqlite
	assert(sq != nil, "This should only be called with non-nil sqlite config")

	if sq.Path == "" {
		return nil, errors.New("Path not specified for database file")
	}

	db, err := sql.Open("sqlite3", sq.Path)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to SQLite database: %s", sq.Path)

	return db, nil
}
