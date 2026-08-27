services:
  chaosbox:
    image: ghcr.io/kive-sh/chaosbox:latest
    restart: unless-stopped
    ports:
      - "{{ get "kive/bucket" "chaosbox_http_port" }}:8080"
