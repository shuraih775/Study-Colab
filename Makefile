.PHONY: up up-core up-all down restart build logs ps clean clean-volumes

# ---------- Core services only ----------
up:
	docker compose up -d jaeger elasticsearch postgres minio file-service core-service prometheus grafana alertmanager

# ---------- All services (including frontend) ----------
up-all:
	docker compose up -d

# ---------- Build images ----------
build:
	docker compose build

# ---------- Stop everything ----------
down:
	docker compose down

# ---------- Restart ----------
restart: down up

# ---------- Logs ----------
logs:
	docker compose logs -f

# ---------- Status ----------
ps:
	docker compose ps

# ---------- Clean containers ----------
clean:
	docker compose down --remove-orphans

# ---------- Clean containers + volumes  ----------
clean-volumes:
	docker compose down -v --remove-orphans
