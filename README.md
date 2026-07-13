# RentNestHub

RentNestHub is a lightweight rental information platform built with a Go API
and a TypeScript/React web client. Landlords can publish listings, tenants can
search with multiple filters, and the recommendation endpoint can rank homes
against natural-language requirements.

## Architecture

```text
.
├── backend/                 Go REST API
│   ├── cmd/api/             application entry point
│   └── internal/
│       ├── config/          environment configuration
│       ├── domain/          core models and repository contracts
│       ├── httpapi/         HTTP routes and handlers
│       ├── repository/      MySQL persistence
│       └── service/         recommendation logic
├── frontend/                Vite + React + TypeScript client
├── deploy/mysql/            schema and development seed data
├── docker-compose.yml       production-like local stack
└── Makefile                 common development commands
```

## Included APIs

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health` | health check |
| `GET` | `/api/v1/houses` | filter listings by region, rent, rooms, or keyword |
| `GET` | `/api/v1/houses/{id}` | listing details |
| `POST` | `/api/v1/houses` | publish a listing with optional image uploads |
| `POST` | `/api/v1/recommendations` | rank listings for tenant requirements |
| `POST` | `/api/v1/favorites` | favorite a listing |
| `DELETE` | `/api/v1/favorites/{tenantId}/{houseId}` | remove a favorite |
| `POST` | `/api/v1/messages` | send an in-site inquiry |

The recommendation service provides deterministic local ranking out of the box.
Set `AI_API_URL`, `AI_API_KEY`, and `AI_MODEL` to reserve configuration for an
external large language model integration.

## Quick Start

1. Copy the environment template:

   ```bash
   cp .env.example .env
   ```

2. Start the complete stack:

   ```bash
   docker compose up --build
   ```

3. Open `http://localhost:8080`. The API is exposed through the same origin at
   `/api`, and MySQL is available on `localhost:3306`.

Uploaded listing images are stored in the `house_uploads` volume. MySQL data is
stored in the `mysql_data` volume.

## Local Development

Backend:

```bash
cd backend
go mod download
go run ./cmd/api
```

Frontend:

```bash
cd frontend
pnpm install
pnpm dev
```

The Vite development server proxies `/api` and `/uploads` to
`http://localhost:8081`.

## Git Workflow

- `main`: stable releases
- `develop`: integration branch
- `feature/house-publish`: listing publication work
- `feature/search`: search and filtering work
- `feature/ai-recommend`: recommendation work

Use Conventional Commits, for example:

```text
feat(search): add district and rent filters
fix(house): validate uploaded image type
docs: update ECS deployment notes
```

Feature branches should be merged into `develop` after review.

## ECS Deployment Notes

See [ECS deployment guide](docs/deployment/ecs.md) for environment setup,
security groups, persistent volume backup, upgrades, and rollback.
