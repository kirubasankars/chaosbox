services:
  chaosbox_a:
    image: ghcr.io/kive-sh/chaosbox:latest
    restart: unless-stopped
    networks:
      - chaosbox_peers
    ports:
      - "{{ get "kive/bucket" "chaosbox_a_http_port" }}:8080"
    volumes:
      - ./config.json:/app/config.json:ro
    command:
      - -config=/app/config.json

networks:
  chaosbox_peers:
    name: chaosbox_peers
    external: true
