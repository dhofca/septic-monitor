package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"sceptic-monitor/internal/db"
	"sceptic-monitor/internal/sms"

	"github.com/joho/godotenv"
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

var (
	lastNotifiedAt  time.Time
	notificationMux sync.Mutex
)

// checkAndNotify checks if level threshold is reached and sends SMS if needed
func checkAndNotify(level float64) {
	notificationMux.Lock()
	defer notificationMux.Unlock()

	// Get threshold from environment
	thresholdStr := os.Getenv("LEVEL_THRESHOLD")
	if thresholdStr == "" {
		return // No threshold configured
	}

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		log.Printf("Invalid LEVEL_THRESHOLD value: %v", err)
		return
	}

	// Check if level has reached or exceeded threshold
	if level < threshold {
		return // Level below threshold, no notification needed
	}

	// Prevent duplicate notifications within 1 hour
	if time.Since(lastNotifiedAt) < time.Hour {
		log.Printf("Notification already sent recently, skipping (level: %.2f, threshold: %.2f)", level, threshold)
		return
	}

	// Send SMS notification
	message := fmt.Sprintf("Alert: Level %.2f has reached the threshold of %.2f", level, threshold)
	if err := sms.Send(message); err != nil {
		log.Printf("Error sending SMS notification: %v", err)
		return
	}

	lastNotifiedAt = time.Now()
	log.Printf("SMS notification sent: level %.2f reached threshold %.2f", level, threshold)
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
	if err := db.SaveLevelData(req.Level); err != nil {
		log.Printf("Error saving to database: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Check if level threshold is reached and send SMS notification
	go checkAndNotify(req.Level)

	// Create response
	response := Response{
		Status:  "success",
		Message: fmt.Sprintf("Received and saved: %f", req.Level),
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleGetLevelData(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get latest level data
	levelData, err := db.GetLatestLevelData()
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
	if err := db.Init(); err != nil {
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
