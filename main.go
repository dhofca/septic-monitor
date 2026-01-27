package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Request represents the incoming POST request body
type Request struct {
	Level float64 `json:"level"`
}

// Response represents the API response
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var db *sql.DB

// initDB initializes the database connection and creates the table
func initDB() error {
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

// saveLevelData saves the level data to the database
func saveLevelData(level float64) error {
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

// handlePOST handles POST requests
func handlePOST(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Save to database
	if err := saveLevelData(req.Level); err != nil {
		log.Printf("Error saving to database: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Create response
	response := Response{
		Status:  "success",
		Message: fmt.Sprintf("Received and saved: %f", req.Level),
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Register the POST endpoint
	http.HandleFunc("/api", handlePOST)

	// Start server on port 8080
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Println("POST endpoint available at: http://localhost:8080/api")
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
