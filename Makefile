COMPOSE := docker compose --env-file .env -f deploy/docker-compose.yml

.PHONY: up down reset logs

up:
	$(COMPOSE) up -d --build
down:
	$(COMPOSE) down
reset:
	$(COMPOSE) down -v
logs:
	$(COMPOSE) logs -f
