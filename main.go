package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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

func handleSaveLevelData(w http.ResponseWriter, r *http.Request) {
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

func getLatestLevelData() (float64, error) {
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

func handleGetLevelData(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get latest level data
	levelData, err := getLatestLevelData()
	if err != nil {
		log.Printf("Error getting level data: %v", err)
		http.Error(w, "Failed to get level data", http.StatusInternalServerError)
		return
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(levelData)
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found.")
	}

	port := os.Getenv("PORT")

	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Register the POST endpoint
	http.HandleFunc("/api", handleSaveLevelData)
	http.HandleFunc("/api/level", handleGetLevelData)

	// Start server
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("POST endpoint available at: http://localhost%s/api\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
