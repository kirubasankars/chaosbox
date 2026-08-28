.PHONY: cert-permissions start stop restart status logs test

COMPOSE ?= docker compose -f docker-compose.yml
CERT_DIR ?= certs
LOG_TAIL ?= 100

include config.env
export

cert-permissions:
	@if [ -d "$(CERT_DIR)" ]; then \
		chmod 0755 "$(CERT_DIR)"; \
		find "$(CERT_DIR)" -type f -name '*.key' -exec chmod 0600 {} +; \
		find "$(CERT_DIR)" -type f ! -name '*.key' -exec chmod 0644 {} +; \
	fi

start: cert-permissions
	$(COMPOSE) up -d --force-recreate --remove-orphans

stop:
	$(COMPOSE) down --remove-orphans

restart: stop start

status:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs --tail=$(LOG_TAIL) --no-color

test:
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -sf "http://127.0.0.1:$(HTTP_PORT)/" >/dev/null && \
		curl -sf "http://127.0.0.1:$(HTTP_PORT)/metrics" | grep -q chaosbox_count && \
		{ echo chaosbox-ok; exit 0; }; \
		sleep 3; \
	done; exit 1
