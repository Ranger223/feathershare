package models

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}
	return DB.Ping()
}
