package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init initializes the database connection and creates the table
func Init() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS level_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// SaveLevelData saves the level data to the database
func SaveLevelData(level float64) error {
	stmt, err := db.Prepare("INSERT INTO level_data (level, created_at) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(level, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	return nil
}

// GetLatestLevelData retrieves the latest level data from the database
func GetLatestLevelData() (float64, error) {
	rows, err := db.Query("SELECT level FROM level_data ORDER BY created_at DESC LIMIT 1")
	if err != nil {
		return 0, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var level float64
		if err := rows.Scan(&level); err != nil {
			return 0, fmt.Errorf("failed to scan level: %w", err)
		}
		return level, nil
	}

	return 0, fmt.Errorf("no level data found")
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
