package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// EventarcPayload represents the structure of the Eventarc JSON payload.
// We'll use an interface{} to handle the dynamic nature of the payload.
type EventarcPayload map[string]interface{}

func main() {
	// Determine the port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	http.HandleFunc("/", eventHandler)

	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)

		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON payload
	var payload EventarcPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Failed to unmarshal JSON payload: %v", err)
		http.Error(w, "Failed to unmarshal JSON payload", http.StatusBadRequest)

		return
	}

	// Pretty print the payload
	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON for pretty printing: %v", err)
		http.Error(w, "Failed to format JSON", http.StatusInternalServerError)

		return
	}

	// Print to the logs
	log.Printf("Received Eventarc Payload:%s", string(prettyJSON))

	// Respond to the request
	fmt.Fprintf(w, "Event received and logged successfully.")
}
