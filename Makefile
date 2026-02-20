COMPOSE_FILE=infra/docker-compose.yml

up:
	docker compose -f $(COMPOSE_FILE) up --build

down:
	docker compose -f $(COMPOSE_FILE) down -v

seed:
	docker compose -f $(COMPOSE_FILE) run --rm migrate

logs:
	docker compose -f $(COMPOSE_FILE) logs -f --tail=200
