.PHONY: build test run up up-all down logs ps clean certs shell-app shell-db

build:
	docker compose build

test:
	cd app && go test -v -race ./...

run:
	cd app && go run .

up:
	docker compose up -d

up-all:
	docker compose --profile monitoring --profile dev up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

clean:
	docker compose down -v --remove-orphans

certs:
	cd nginx/certs && bash generate-certs.sh

shell-app:
	docker compose exec app sh

shell-db:
	docker compose exec postgres psql -U $${POSTGRES_USER:-app} -d $${POSTGRES_DB:-compose_stack}
