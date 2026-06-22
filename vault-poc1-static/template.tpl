{{ with secret "gcp/static-account/my-app-key-rotator/key" }}
{{ base64Decode .Data.private_key_data }}
{{ end }}