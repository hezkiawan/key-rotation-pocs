package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Credentials matches the structure of the JSON key provided by GCP/Vault.
type Credentials struct {
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
}

var (
	activeCreds Credentials
	bootTime    time.Time
)

func main() {
	bootTime = time.Now()
	fmt.Println("Starting application boot sequence...")

	// =========================================================================
	// POC ARCHITECTURE CONCEPT: The "Dumb" Application
	// Notice there is zero rotation logic, no Mutexes, and no SIGHUP listeners.
	// The application assumes the infrastructure will provide the file on boot
	// and relies on Kubernetes to kill the app when the file changes.
	// =========================================================================

	// 1. Read the file EXACTLY ONCE at startup.
	file, err := os.ReadFile("/app/secrets/firebase-sa.json")
	if err != nil {
		// In Kubernetes, failing here is a feature, not a bug
		// If the secret isn't mounted yet, the Pod will 'CrashLoopBackOff'
		// and try again automatically until Vault Secrets Operator delivers it.
		panic(fmt.Sprintf("Fatal: Could not read secret file: %v", err))
	}

	// 2. Parse the JSON payload
	if err := json.Unmarshal(file, &activeCreds); err != nil {
		panic(fmt.Sprintf("Fatal: Error parsing JSON: %v", err))
	}

	fmt.Printf("✅ Credentials Loaded on Boot! Key ID: %s\n", activeCreds.PrivateKeyID[:10]+"...")

	// 3. Serve the data to prove zero-downtime rotation
	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":                 "success",
			"active_service_account": activeCreds.ClientEmail,
			"private_key_id":         activeCreds.PrivateKeyID,
			"pod_booted_at":          bootTime.Format(time.RFC3339),
		})
	})

	fmt.Println("🚀 Web server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
