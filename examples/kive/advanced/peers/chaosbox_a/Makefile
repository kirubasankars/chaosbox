.PHONY: cert-permissions start stop restart status logs peers-net

COMPOSE ?= docker compose -f docker-compose.yml
CERT_DIR ?= certs
LOG_TAIL ?= 100

cert-permissions:
	@if [ -d "$(CERT_DIR)" ]; then \
		chmod 0755 "$(CERT_DIR)"; \
		find "$(CERT_DIR)" -type f -name '*.key' -exec chmod 0600 {} +; \
		find "$(CERT_DIR)" -type f ! -name '*.key' -exec chmod 0644 {} +; \
	fi

peers-net:
	@docker network inspect chaosbox_peers >/dev/null 2>&1 || docker network create chaosbox_peers

start: cert-permissions peers-net
	$(COMPOSE) up -d --force-recreate --remove-orphans

stop:
	$(COMPOSE) down --remove-orphans

restart: stop start

status:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs --tail=$(LOG_TAIL) --no-color
