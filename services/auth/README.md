# Auth Service (Go)

RideHail auth — Register, Login, Refresh, JWT, PostgreSQL (PostGIS + migrations), Redis (optional), OAuth stubs.

## Run locally

1. Start infra: `docker compose -f ../../infra/docker-compose.yml up -d postgres redis`
2. `go mod tidy && go run .`
3. Health: `curl http://localhost:8080/health`
4. Register: `POST http://localhost:8080/auth/register` — `{"email":"u@example.com","password":"password123","role":"passenger"}`
5. Login: `POST http://localhost:8080/auth/login` — `{"email":"u@example.com","password":"password123"}`
6. Refresh: `POST http://localhost:8080/auth/refresh` — `{"refresh_token":"..."}`
7. OAuth stubs: `GET /auth/oauth/google`, `/auth/oauth/yandex`, `/auth/oauth/vk` — 501

## Env

- `PORT` (default 8080)
- `PG_DSN`
- `JWT_SECRET` (change in production)
- `REDIS_ADDR` (optional; empty = no Redis)
