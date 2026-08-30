services:
  chaosbox:
    image: ghcr.io/kive-sh/chaosbox:latest
    pull_policy: always
    restart: unless-stopped
    ports:
      - "{{ get "kive/bucket" "chaosbox_http_port" }}:8080"
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
