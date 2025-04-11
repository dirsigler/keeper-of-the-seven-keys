package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// EventarcPayload represents the structure of the Eventarc JSON payload.
type EventarcPayload map[string]interface{}

func main() {
	// Configure slog to output JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Determine the port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		slog.Debug("Defaulting to port", "port", port)
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
	slog.Debug("Received Eventarc Payload", "payload", payload)

	// Extract relevant information from the payload
	cryptoKeyName, ok := extractCryptoKeyName(payload)
	if !ok {
		slog.Error("Failed to extract crypto key name from payload")
		http.Error(w, "Failed to extract crypto key name from payload", http.StatusBadRequest)

		return
	}

	// Update the rotation period
	if err := updateCryptoKeyRotationPeriod(cryptoKeyName); err != nil {
		slog.Error("Failed to update crypto key rotation period", "error", err, "cryptoKeyName", cryptoKeyName)
		http.Error(w, "Failed to update crypto key rotation period", http.StatusInternalServerError)

		return
	}

	// Respond to the request
	fmt.Fprintf(w, "Event received and processed successfully.")
}

// extractCryptoKeyName extracts the crypto key name from the Eventarc payload.
func extractCryptoKeyName(payload EventarcPayload) (string, bool) {
	protoPayload, ok := payload["protoPayload"].(map[string]interface{})
	if !ok {
		return "", false
	}

	resourceName, ok := protoPayload["resourceName"].(string)
	if !ok {
		return "", false
	}

	return resourceName, true
}

// updateCryptoKeyRotationPeriod updates the rotation period of a crypto key to 90 days.
func updateCryptoKeyRotationPeriod(cryptoKeyName string) error {
	ctx := context.Background()

	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create kms client: %v", err)
	}

	defer client.Close()

	// Set the new rotation period to 90 days
	rotationPeriod := durationpb.New(90 * 24 * time.Hour)

	// Build the update request
	req := &kmspb.UpdateCryptoKeyRequest{
		CryptoKey: &kmspb.CryptoKey{
			Name: cryptoKeyName,
			RotationSchedule: &kmspb.CryptoKey_RotationPeriod{
				RotationPeriod: rotationPeriod,
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"rotation_period"},
		},
	}

	// Update the crypto key
	_, err = client.UpdateCryptoKey(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update crypto key: %v", err)
	}

	slog.Info("Successfully updated crypto key rotation period", "cryptoKeyName", cryptoKeyName, "rotationPeriod", rotationPeriod.AsDuration().String())

	return nil
}
