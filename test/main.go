package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const lokiAddress = "http://172.16.16.4:32401/loki/api/v1/push"

func sendToLoki(logMessage string, streamLabels map[string]string) error {
	// Create a timestamp in nanoseconds
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	// Construct the payload
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": streamLabels,
				"values": [][]string{
					{timestamp, logMessage},
				},
			},
		},
	}

	// Marshal the payload to JSON
	reqJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Make the POST request to Loki
	resp, err := http.Post(lokiAddress, "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("error sending log to Loki: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Check if the request was successful
	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to send log, status: %s, body: %s", resp.Status, body)
	}

	fmt.Printf("Log sent successfully: %s\n", body)
	return nil
}

func main() {
	// Define the log message and labels
	logMessage := "fizzbuzz"
	streamLabels := map[string]string{
		"foo": "bar2",
	}

	// Send the log to Loki
	if err := sendToLoki(logMessage, streamLabels); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
