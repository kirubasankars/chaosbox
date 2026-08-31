{
  "version": "0.1.0",
  "listen": ":8080",
  "tls_cert": "",
  "tls_key": "",
  "peers": [
{{- $port := get "kive/bucket" "chaosbox_http_port" }}
{{- range $i, $ip := split (getJobWorkerOptional "peer_workers") "," }}{{ if $ip }}{{ if $i }},{{ end }}
    "{{ $ip }}:{{ $port }}"{{ end }}{{ end }}
  ],
  "peer_check_sec": 5,
  "peer_ca_cert": ""
}
