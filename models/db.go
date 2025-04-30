package models

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}

	fmt.Println("Connected to DB")

	return migrate()
}

func migrate() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`

	_, err := DB.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	fmt.Println("Users table ready")

	sessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	_, err = DB.Exec(sessionsTable)
	if err != nil {
		return fmt.Errorf("error creating sessions table: %w", err)
	}

	filesTable := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		filename TEXT NOT NULL,
		filepath TEXT NOT NULL,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	_, err = DB.Exec(filesTable)
	if err != nil {
		return fmt.Errorf("error creating files table: %w", err)
	}

	return nil
}
