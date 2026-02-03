# User Service (Go)

RideHail User — profile CRUD, driver verification stub, JWT-protected API.

## Run locally

1. Start infra (Postgres + Redis): `docker compose -f ../../infra/docker-compose.yml up -d postgres redis`
2. Run Auth first (creates users + migrations for profiles): `cd ../auth && go run .`
3. `go mod tidy && go run .`
4. Get token from Auth: `POST http://localhost:8080/auth/login` or `/auth/register`
5. `GET/PATCH http://localhost:8081/api/v1/users/me` with `Authorization: Bearer <access_token>`
6. `POST http://localhost:8081/api/v1/users/me/driver` with `{"license_number":"..."}` — driver verification stub (doc upload later)

## Env

- `PORT` (default 8081)
- `PG_DSN` (same as Auth)
- `REDIS_ADDR` (optional)
- `JWT_SECRET` (must match Auth)
