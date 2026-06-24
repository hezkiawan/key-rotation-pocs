# Phase 3: Workload Identity Federation (Zero-Trust)

This Proof of Concept demonstrates the ultimate evolution of cloud secret management: **Keyless Authentication**. By utilizing Workload Identity Federation (WIF) and OpenID Connect (OIDC), this architecture completely eliminates the need for static Service Account Keys (`.json` files).

## Architecture Overview

Instead of storing a permanent password in a CI/CD pipeline or a Vault server, this approach relies on cryptographic trust established between Google Cloud and an Identity Provider (in this case, GitHub Actions).

1. **Identity Provider (GitHub):** GitHub generates a short-lived, cryptographically signed OIDC token containing claims about the repository and branch.
2. **The WIF Trust Boundary (GCP):** Google Cloud receives the token, verifies GitHub's signature, and evaluates the token's claims against strict Authorization Conditions.
3. **Ephemeral Token Exchange:** If conditions pass, GCP issues a temporary (1-hour) access token bound specifically to the requested Service Account.
4. **Application Default Credentials (ADC):** The Go application relies entirely on ADC (`firebase.NewApp(ctx, nil)`). It automatically detects the injected environment token and authenticates without any credential-handling code.

## GCP Console Setup (The Trust Boundary)

To make this architecture work, the "locks" must be configured inside the Google Cloud Console. 

### 1. Create the WIF Pool
The Pool acts as the isolated logical boundary for external identities.
* Navigate to **IAM & Admin > Workload Identity Federation**.
* Click **Create Pool** and name it (e.g., `github-pool`).

### 2. Create the WIF Provider
The Provider translates the external GitHub token into Google-readable attributes.
* Inside the Pool, click **Add Provider** (Select OpenID Connect / GitHub).
* **Issuer URL:** `https://token.actions.githubusercontent.com`
* **Attribute Mapping:** * `google.subject` = `assertion.sub`
  * `attribute.repository` = `assertion.repository`
* **Attribute Conditions (Optional but Recommended):** * `attribute.repository == "your-username/key-rotation-pocs"`

### 3. Allow Impersonation (IAM Binding)
Grant the external GitHub identity permission to assume the roles of your target GCP Service Account.
* In the Pool UI, click **Grant Access**.
* Select your target Service Account (e.g., `firebase-app-emulator@...`).
* Define the principal set to restrict access to your specific repository.

# How to Run the PoC

Because this architecture relies on the environment rather than static files, there are two distinct ways to execute the backend.

### Method 1: CI/CD Pipeline (GitHub Actions)
This demonstrates the true WIF OIDC handshake in a production-like automated environment.
1. Update the `workload_identity_provider` and `service_account` strings in `.github/workflows/deploy.yml` with your GCP Project details.
2. Push the code to the `main` branch.
3. Open the **Actions** tab in GitHub to watch the runner successfully trade the OIDC token, inject it into the environment, and execute the Go application natively.

### Method 2: Local Developer Experience (DX)
To run this application locally without downloading a static master key:
1. Ensure the Google Cloud CLI (`gcloud`) is installed on your machine.
2. Run the ADC login command to generate a local developer token:
```bash
gcloud auth application-default login

Run the Go application:

Bash 
go run main.go

Result: The SDK will automatically detect your local developer identity and authenticate to Firebase.