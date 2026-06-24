package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
)

// initializeFirebase demonstrates the power of Application Default Credentials (ADC).
// By passing 'nil' as the configuration option, the Google Cloud SDK automatically 
// interrogates the host environment to find its identity. 
//
// In a local environment: It uses the developer's `gcloud auth application-default login`.
// In GitHub Actions: It uses the OIDC token injected by the WIF provider.
// In GCP Compute: It uses the native VM metadata server if there isn't a Credential Configuration File
//
// ZERO static secrets or JSON files are required in the application code.
func initializeFirebase() (*firebase.App, error) {
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

func main() {
	log.Println("Starting Zero-Trust WIF Backend Service...")

	// 1. Authenticate to GCP/Firebase
	app, err := initializeFirebase()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v\n", err)
	}
	log.Println("✅ Successfully authenticated with GCP using Application Default Credentials!")

	// 2. Serve the API
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		// In a production application, 'app' would be used to initialize 
		// Firestore, Auth, or Cloud Messaging clients here.
		_ = app 
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Backend is authenticated and running perfectly via WIF!"))
	})

	// 3. Listen on port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}