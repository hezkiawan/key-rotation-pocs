# Vault PoC 2: Zero-Trust Kubernetes Secret Management

This Proof of Concept demonstrates an enterprise-grade, zero-trust architecture for managing and automatically rotating secrets in Kubernetes. 

Instead of hardcoding credentials or manually managing Kubernetes Secrets, this architecture forces the application cluster to authenticate with an external security authority (HashiCorp Vault) to fetch short-lived, dynamically generated credentials.

## 🏗️ Architecture Overview

* **The Security Authority:** HashiCorp Vault runs natively on the host machine (simulating a centralized, external enterprise Vault cluster).
* **The Bridge (VSO):** The HashiCorp Vault Secrets Operator runs inside Kubernetes. It authenticates with Vault using a native Kubernetes Service Account, fetches a Base64-encoded GCP credential, decodes it, and mounts it as a physical JSON file.
* **The Zero-Downtime Engine:** The `stakater/Reloader` operator watches the mounted secret. When the 2-minute Vault lease expires and VSO updates the file, Reloader triggers a seamless Kubernetes Rolling Update.
* **The Application:** A lightweight Go backend that contains *zero* secret-rotation logic. It reads the file once on boot and relies entirely on the infrastructure for security.

## 📋 Prerequisites

To run this PoC locally, you will need:
1. **Windows OS** with **Git Bash** (MINGW64).
2. **Docker Desktop** & **Minikube** installed and running.
3. **HashiCorp Vault** (`vault.exe`) downloaded and placed in the root directory.
4. A valid GCP Service Account JSON key placed at `app/master-sa-key.json` (This is the "master" key Vault uses to dynamically generate temporary child keys).

---

## 🚀 How to Run the PoC

Open a Git Bash terminal in the root directory and execute the following steps:

**Step 1: Start the Kubernetes Cluster**
Ensure your local container engine is running, then boot Minikube:
```bash
minikube start
```

**Step 2: Boot Vault & Establish Trust**
Run the automated setup script. This will start an external Vault server in the background, configure the GCP Secrets Engine with a 2-minute TTL, and establish a cryptographic trust bridge with your Minikube cluster.
```bash
./setup-vault.sh
```

**Step 3: Deploy the Infrastructure**
Apply the Kubernetes manifests. This will create the Service Account, map the Vault connection, and spin up the Go backend.
```bash
kubectl apply -f k8s/
```
*Wait until the application is fully running:* `kubectl get pods -n app-namespace -w`

---

## 🧪 The Zero-Downtime Rotation Test

To prove the architecture works, we will simulate a production traffic load and watch the infrastructure rotate the secret dynamically.

**1. Expose the Application behind a stable Service:**
```bash
kubectl expose deployment go-backend --name=go-service --type=NodePort --port=8080 -n app-namespace
```

**2. Get the local Minikube URL:**
```bash
minikube service go-service -n app-namespace --url
```
*(Copy the URL output from this command, e.g., `http://127.0.0.1:53214`)*

**3. Run the Traffic Loop:**
Replace `<YOUR_URL>` below with the URL from the previous step and run this loop in your terminal:
```bash
while true; do curl -s <YOUR_URL>/api/data; echo ""; sleep 1; done
```

**What to Watch For:**
Leave the loop running. Because the Vault lease is set to exactly 2 minutes, you will see the API successfully returning data. Right at the 2-minute mark, the `private_key_id` will instantly swap to a brand new key, and the `pod_booted_at` timestamp will update. **Notice that not a single HTTP connection is dropped during the rotation.**

---

## 🧹 Cleanup

When you are finished testing, cleanly tear down the local infrastructure:

```bash
# 1. Delete the Kubernetes resources
kubectl delete -f k8s/

# 2. Kill the background Vault process
taskkill //F //IM vault.exe
```