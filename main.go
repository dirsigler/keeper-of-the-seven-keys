package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

// EventarcPayload represents the structure of the Eventarc JSON payload.
// We'll use an interface{} to handle the dynamic nature of the payload.
type EventarcPayload map[string]interface{}

func init() {
	functions.HTTP("eventHandler", eventHandler)
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	// Decode the JSON payload
	var payload EventarcPayload

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("Failed to decode JSON payload: %v", err)
		http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)

		return
	}

	defer r.Body.Close()

	// Pretty print the payload
	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON for pretty printing: %v", err)
		http.Error(w, "Failed to format JSON", http.StatusInternalServerError)

		return
	}

	// Print to the logs
	log.Printf("Received Eventarc Payload:\n%s", string(prettyJSON))

	// Respond to the request
	fmt.Fprintf(w, "Event received and logged successfully.")
}
