.PHONY: dev up down logs test

dev:
	docker compose up --build

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

test:
	cd backend && go test ./...
	cd frontend && pnpm lint && pnpm build
