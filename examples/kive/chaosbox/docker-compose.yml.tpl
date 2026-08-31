services:
  chaosbox:
    image: ghcr.io/kive-sh/chaosbox:latest
    pull_policy: always
    # Match coroot_node_agent --container-allowlist=^/docker/<bucket_id>-
    # (Compose's default {project}-{service}-1 name is skipped).
    container_name: "{{ .BucketID }}-{{ .Job }}"
    restart: unless-stopped
    ports:
      - "{{ get "kive/bucket" "chaosbox_http_port" }}:8080"
    volumes:
      - ./config.json:/app/config.json:ro
    command:
      - -config=/app/config.json
    # json-file so coroot_node_agent can tail stdout (local driver is not supported).
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
    labels:
      kive.bucket: "{{ .BucketID }}"
      kive.job: "{{ .Job }}"
      kive.allocation: "{{ .AllocationID }}"
