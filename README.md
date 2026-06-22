# Service Account Key Rotation & Authentication PoCs

This repository contains a series of Proof of Concept (PoC) implementations exploring the evolution of cloud authentication and secret management. The objective is to demonstrate the architectural progression from managing static long-lived credentials to implementing dynamic rotation, and ultimately achieving a completely keyless zero-trust environment.

All backend applications in this repository are written in **Go** and designed to authenticate securely with Google Cloud Platform (GCP).

## Architecture Evolution

This project is divided into three distinct phases. Each phase solves the vulnerabilities of the previous one, culminating in modern, enterprise-grade secret management.

### 1. [Phase 1: Bare VM Dynamic Rotation (Vault Agent)](./vault-poc1-static)
Demonstrates how to achieve zero-downtime credential hot-swapping in a standard operating system environment without an orchestrator.
* **Core Technology:** HashiCorp Vault, Vault Agent, OS Signaling (`SIGHUP`).
* **Achievement:** Replaced permanent static keys with short-lived (1-minute) dynamically generated keys that are automatically rotated and injected into the Go backend's memory.

### 2. [Phase 2: Kubernetes Dynamic Secret Rotation](./vault-poc2-dynamic)
Scales the dynamic rotation concept into a container orchestration environment, utilizing Kubernetes native mechanisms to distribute secrets securely to pods.
* **Core Technology:** HashiCorp Vault, Vault Secrets Operator (VSO), Kubernetes Service Accounts.
* **Achievement:** Eliminates the need to manage background daemons manually. Vault natively synchronizes temporary credentials directly into the Kubernetes pod filesystem.

### 3. [Phase 3: Workload Identity Federation (Zero-Trust)](./wif-poc-zerotrust)
The final architecture completely eliminates HashiCorp Vault, secret managers, and physical `.json` keys. It establishes a strict zero-trust boundary based on temporary cryptographic proof.
* **Core Technology:** Google Cloud Workload Identity Federation (WIF), OpenID Connect (OIDC), Application Default Credentials (ADC), GitHub Actions.
* **Achievement:** **Keyless Authentication.** The application and deployment pipelines authenticate natively via their runtime environment context. No secrets are ever downloaded, stored, or managed.

## Getting Started

To explore a specific architecture, navigate to the respective directory. Each phase contains:
1. A standalone Go backend application.
2. The necessary infrastructure configuration (Dockerfiles, Kubernetes manifests, or GitHub Actions workflows).
3. A dedicated `README.md` detailing the architectural theory and execution steps for that specific phase.

## Security Note
To ensure the integrity of these architectures, **no GCP Service Account Keys are committed to this repository**. When running Phase 1 or Phase 2 locally, you must provide your own `master-sa-key.json` securely as outlined in their respective documentations. Phase 3 operates entirely without keys.