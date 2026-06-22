#!/bin/bash
set -e

echo "1. Starting Vault Server (Dev Mode)"
vault server -dev -dev-root-token-id="root" -dev-listen-address="0.0.0.0:8200" &
VAULT_PID=$!
sleep 3

export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="root"

echo "2. Configuring Vault GCP Secrets Engine"
vault secrets enable gcp
vault secrets tune -default-lease-ttl=1m -max-lease-ttl=1m gcp

# Bind the master key (This gives Vault permission to create short-lived keys)
vault write gcp/config \
    credentials=@/app/master-sa-key.json \
    project="vault-rotation-sandbox"

# Configure the specific service account to rotate
vault write gcp/static-account/my-app-key-rotator \
    service_account_email="firebase-app-emulator@vault-rotation-sandbox.iam.gserviceaccount.com" \
    secret_type="service_account_key"

echo "3. Creating Policy and AppRole for Vault Agent"
# Give agent permisson to ONLY read this exact path
vault policy write gcp-reader - <<EOF
path "gcp/static-account/my-app-key-rotator/key" {
  capabilities = ["read"]
}
EOF

vault auth enable approle
vault write auth/approle/role/vault-agent token_policies="gcp-reader"

# Save the Machine Identity (RoleID and SecretID) for the Vault Agent to authenticate to Vault
vault read -field=role_id auth/approle/role/vault-agent/role-id > /app/role_id
vault write -field=secret_id -f auth/approle/role/vault-agent/secret-id > /app/secret_id

echo "4. Building and launching the Go Backend"
go build -o /app/main /app/main.go
/app/main &
GO_APP_PID=$!
sleep 2

echo "5. Starting Vault Agent Daemon..."
# This process runs continuously, authenticating via AppRole and writing the template
vault agent -config=/app/vault-agent.hcl