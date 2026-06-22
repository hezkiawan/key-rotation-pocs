# Phase 1: Bare VM Dynamic Rotation (Vault Agent)

This Proof of Concept demonstrates how to achieve dynamic, zero-downtime Service Account Key rotation on a standard OS-level architecture (e.g., a bare Ubuntu Virtual Machine).

## The Architecture Baseline

Deploying to a bare virtual machine requires handling infrastructure at the raw operating system level. Because the application runs directly on the host, the deployment strategy must account for explicit system management:

* **Direct File Management:** Credential distribution relies on strict file system permissions rather than abstracted secret stores.
* **Process Management:** Service lifecycles and background daemons require direct OS-level configuration and monitoring.
* **Internal State Handling:** The application itself holds the responsibility for dynamic state management, requiring built-in logic to detect and reload rotated credentials without external restarts.

## Implementation Overview

To meet these baseline requirements, this architecture utilizes **Vault Agent** as a background daemon and native OS signals (`SIGHUP`) to achieve application hot-swapping.

1. **Vault Engine:** Acts as the central cryptographic authority, securely holding the GCP Master Key and dynamically generating short-lived (1-minute) Service Account Keys.
2. **Vault Agent (Daemon):** Runs continuously on the VM. It authenticates with Vault via AppRole, fetches the temporary GCP key, and writes it to the local file system.
3. **OS Signaling:** When a key is about to expire, Vault Agent fetches a new one, rewrites the file, and executes an OS command (`pkill -SIGHUP main`) to notify the application.
4. **Go Application:** A lightweight backend that listens for the `SIGHUP` interrupt. Upon receiving the signal, it safely locks memory (`sync.RWMutex`) and reloads the new JSON file without dropping active user requests.

## How to Run the PoC

This simulation packages the Vault server, the Go application, and the Vault Agent into a single Ubuntu Docker container to accurately mimic a self-contained Virtual Machine environment.

### Prerequisites
* You must place a valid GCP Service Account Key named `master-sa-key.json` in this directory. **(Do not commit this file to Git!)** this Service Account is for the Vault so it must have permission to act as key administrator in the GCP settings.

### Execution Steps

1. **Build the Environment:**
```bash
docker build -t vault-bare-vm-poc .

Start the Simulation:

Bash 
docker run -p 8080:8080 vault-bare-vm-poc

Observe the Dynamic Rotation:

Open a new terminal and query the Go application. Notice the private_key_id.

Bash 
curl http://localhost:8080/api/data

Wait 60 seconds for the Vault Agent to rotate the key. Run the curl command again. You will see the application seamlessly serving data using a completely new private_key_id, proving the background hot-swap was successful.