package common

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB opens a SQLite database and returns the connection.
func OpenDB(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}
