FROM ubuntu:22.04

# Install basic system tools required for OS-level management
# procps is required for the 'pkill' command used by Vault Agent to send SIGHUP
RUN apt-get update && apt-get install -y \
    curl \
    gpg \
    coreutils \
    procps \
    golang-go \
    && rm -rf /var/lib/apt/lists/*

# Install the official HashiCorp Vault binary
RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com jammy main" | tee /etc/apt/sources.list.d/hashicorp.list \
    && apt-get update && apt-get install -y vault

WORKDIR /app
RUN mkdir -p /app/secrets

# Copy application and configuration files
COPY main.go /app/main.go
COPY template.tpl /app/template.tpl
COPY vault-agent.hcl /app/vault-agent.hcl
COPY start.sh /app/start.sh

# IMPORTANT: master-sa-key.json is the Vault admin service account key downloaded from GCP that can control other service account keys
COPY master-sa-key.json /app/master-sa-key.json 

RUN chmod +x /app/start.sh

EXPOSE 8080

CMD ["/app/start.sh"]