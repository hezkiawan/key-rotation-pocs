exit_after_auth = false
pid_file        = "/var/run/vault-agent.pid"

vault {
  address = "http://127.0.0.1:8200"
}

# Machine Authentication: The Agent logs in automatically using AppRole
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path                   = "/app/role_id"
      secret_id_file_path                 = "/app/secret_id"
      remove_secret_id_file_after_reading = false
    }
  }
}

# Rendering Engine: Fetches the GCP key, decodes it, writes it, and triggers the app
template {
  source      = "/app/template.tpl"
  destination = "/app/secrets/firebase-sa.json"
  # OS-level signaling: Sends SIGHUP to our Go binary so it knows to reload the file
  command     = "pkill -SIGHUP main || true"
}