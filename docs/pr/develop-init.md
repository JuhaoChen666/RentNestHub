# PR Title

feat(init): scaffold RentNestHub Go and React architecture

# PR Body

## Summary

Initialize RentNestHub with a Go REST API, TypeScript/React frontend, MySQL schema, and Docker Compose deployment path.

## Scope

- Branch: `develop`
- Target: `main`
- Change size: initial scaffold; exceeds 300 lines because it creates the baseline application structure, lockfiles, and deployment files in one pass.

## Changes

- Add Go backend with health, house publishing, house search, recommendations, favorites, and messages APIs.
- Add MySQL schema and seed data for users, houses, favorites, and messages.
- Add React TypeScript frontend for browsing, filtering, recommending, publishing, favoriting, and messaging.
- Add Docker Compose stack with MySQL, API, frontend, image uploads volume, and Nginx reverse proxy.
- Add environment template, Makefile, and project documentation.
- Add PR/commit workflow guidance for future feature branches.

## Validation

- `go test ./...`
- `pnpm lint`
- `pnpm build`
- `docker compose config --quiet`

## Risk

- First scaffold is intentionally larger than normal PR size guidance.
- AI recommendation currently uses deterministic local ranking; external model integration is only reserved through environment configuration.
- Docker Desktop was not required locally; Docker CLI was validated through Colima.

## Follow-Up PRs

- `feature/house-publish`: split publishing validation, image handling, and landlord workflow into atomic PRs.
- `feature/search`: split filters, pagination, and result sorting into atomic PRs.
- `feature/ai-recommend`: split prompt/model integration, ranking persistence, and recommendation UX into atomic PRs.

