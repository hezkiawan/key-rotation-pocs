package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Credentials tracks the active GCP Service Account Key
// PrivateKeyID is included to explicitly prove the key is dynamically rotating
type Credentials struct {
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
}

var (
	activeCreds Credentials
	lastReload  time.Time
	// mu ensures thread-safe memory updates during a hot-swap rotation
	mu sync.RWMutex
)

// loadCredentials reads the JSON file injected by Vault Agent into memory
func loadCredentials() {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.ReadFile("/app/secrets/firebase-sa.json")
	if err != nil {
		fmt.Println("Warning: Could not read secret file yet. Awaiting Vault Agent...")
		return
	}

	var creds Credentials
	if err := json.Unmarshal(file, &creds); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	activeCreds = creds
	lastReload = time.Now()
	fmt.Printf("✅ Credentials Loaded! New Key ID: %s\n", activeCreds.PrivateKeyID[:10]+"...")
}

func main() {
	// 1. Attempt initial load on boot
	loadCredentials()

	// 2. OS-Level Signal Listener (The Bare VM Method)
	// Listens for SIGHUP sent by the Vault Agent when a new key is written
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)

	go func() {
		for {
			<-sigs
			fmt.Println("🔄 SIGHUP Received: Vault Agent triggered rotation! Reloading into memory...")
			loadCredentials()
		}
	}()

	// 3. Endpoint to serve the current active identity state
	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":                 "success",
			"active_service_account": activeCreds.ClientEmail,
			"private_key_id":         activeCreds.PrivateKeyID,
			"last_rotated_at":        lastReload.Format(time.RFC3339),
		})
	})

	fmt.Println("🚀 Web server listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}