package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// EventarcPayload represents the structure of the Eventarc JSON payload.
// We'll use an interface{} to handle the dynamic nature of the payload.
type EventarcPayload map[string]interface{}

func main() {
	// Configure slog to output JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Determine the port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		slog.Info("Defaulting to port", "port", port)
	}

	http.HandleFunc("/", eventHandler)

	slog.Info("Listening on port", "port", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)

		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON payload
	var payload EventarcPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("Failed to unmarshal JSON payload", "error", err)
		http.Error(w, "Failed to unmarshal JSON payload", http.StatusBadRequest)

		return
	}

	// Log the payload directly as structured data
	slog.Info("Received Eventarc Payload", "payload", payload)

	// Respond to the request
	fmt.Fprintf(w, "Event received and logged successfully.")
}
