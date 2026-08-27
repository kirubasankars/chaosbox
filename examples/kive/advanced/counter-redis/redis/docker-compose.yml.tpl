services:
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    ports:
      - "{{ get "kive/bucket" "chaosbox_redis_port" }}:6379"
