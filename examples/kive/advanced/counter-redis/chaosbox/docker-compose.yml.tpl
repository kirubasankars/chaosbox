services:
  chaosbox:
    image: ghcr.io/kive-sh/chaosbox:latest
    restart: unless-stopped
    extra_hosts:
      - "host.docker.internal:host-gateway"
    ports:
      - "{{ get "kive/bucket" "chaosbox_http_port" }}:8080"
    command:
      - -redis=redis://host.docker.internal:{{ get "kive/bucket" "chaosbox_redis_port" }}
