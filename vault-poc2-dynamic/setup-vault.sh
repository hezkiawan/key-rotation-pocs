#!/bin/bash
set -e

echo "--- Vault PoC 2: Zero-Trust Infrastructure Setup ---"

echo "[Step 0] Cleaning up old Vault processes (Idempotency Check)..."
taskkill //F //IM vault.exe > /dev/null 2>&1 || true
sleep 2 

echo "[Step 1] Starting external Vault Server in Dev Mode..."
./vault.exe server -dev -dev-root-token-id="root" -dev-listen-address="0.0.0.0:8200" > /dev/null 2>&1 &
sleep 3
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="root"
echo " -> Vault running natively on $VAULT_ADDR"

echo "[Step 2] Enabling & Configuring GCP Secrets Engine..."
./vault.exe secrets enable gcp > /dev/null 2>&1 || true

# We force a 2-minute TTL specifically for this PoC to demonstrate rapid rotation
./vault.exe secrets tune -default-lease-ttl=2m -max-lease-ttl=2m gcp > /dev/null 2>&1

./vault.exe write gcp/config \
    credentials=@./app/master-sa-key.json \
    project="vault-rotation-sandbox" > /dev/null 2>&1

./vault.exe write gcp/static-account/my-app-key-rotator \
    service_account_email="firebase-app-emulator@vault-rotation-sandbox.iam.gserviceaccount.com" \
    secret_type="service_account_key" > /dev/null 2>&1
echo " -> GCP engine configured with 2-minute lease TTL."

echo "[Step 3] Creating the Least-Privilege Policy ('gcp-reader')..."
./vault.exe policy write gcp-reader - <<EOF > /dev/null 2>&1
path "gcp/static-account/my-app-key-rotator/key" {
  capabilities = ["read"]
}
EOF

echo "[Step 4] Enabling Kubernetes Authentication & Fetching Minikube Certificates..."
./vault.exe auth enable kubernetes > /dev/null 2>&1 || true

# Extract Minikube's actual external host IP and CA Certificate
K8S_HOST=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.server}')
kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 -d > minikube-ca.crt

# Configure Vault to connect externally to Minikube
./vault.exe write auth/kubernetes/config \
    kubernetes_host="$K8S_HOST" \
    kubernetes_ca_cert=@minikube-ca.crt \
    disable_local_ca_jwt=true > /dev/null 2>&1
echo " -> Trust established with Minikube at $K8S_HOST"

echo "[Step 5] Creating the Trust Bridge (Vault Role -> K8s Service Account)..."
./vault.exe write auth/kubernetes/role/vso-role \
    bound_service_account_names="backend-ksa" \
    bound_service_account_namespaces="app-namespace" \
    token_policies="gcp-reader" \
    ttl=1h > /dev/null 2>&1

echo "✅ Setup Complete: Vault is configured for Zero-Password Kubernetes Auth!"