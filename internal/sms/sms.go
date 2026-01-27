package sms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func Send(message string) error {
	// Get configuration from environment variables
	apiKey := os.Getenv("SMS_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("SMS_API_KEY not configured")
	}

	phoneNumber := os.Getenv("SMS_PHONE_NUMBER")
	if phoneNumber == "" {
		return fmt.Errorf("phone number not configured")
	}

	// Get sender name from environment, default to "Test" if not set
	senderName := os.Getenv("SMS_FROM")
	if senderName == "" {
		senderName = "Test"
	}

	// Prepare URL with parameters
	apiURL := "https://api.smsapi.pl/sms.do"
	params := url.Values{}
	params.Set("to", phoneNumber)
	params.Set("message", message)
	params.Set("from", senderName)
	params.Set("format", "json")

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response to check for errors
	var apiResponse struct {
		Error   int    `json:"error"`
		Message string `json:"message"`
		Count   int    `json:"count"`
		List    []struct {
			ID     string  `json:"id"`
			Points float64 `json:"points"`
		} `json:"list"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// If response is not JSON or doesn't match expected format, log it but don't fail
		log.Printf("SMS API response: %s", string(body))
	}

	if apiResponse.Error != 0 {
		return fmt.Errorf("SMS API error %d: %s", apiResponse.Error, apiResponse.Message)
	}

	if len(apiResponse.List) > 0 {
		log.Printf("SMS sent successfully. Message ID: %s, Points: %.2f", apiResponse.List[0].ID, apiResponse.List[0].Points)
	} else {
		log.Printf("SMS sent successfully. Response: %s", string(body))
	}

	return nil
}
