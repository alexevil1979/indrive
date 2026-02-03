# Payment Service (Go)

Checkout flow, cash/card stub (later: Tinkoff, Sber, YooMoney).

## Run locally

1. Start Postgres: `docker compose -f ../../infra/docker-compose.yml up -d postgres`
2. Run Auth first (same DB)
3. `go mod tidy && go run .`
4. All endpoints require `Authorization: Bearer <token>` (same JWT as Auth).

## API

- **POST /api/v1/payments** — create payment (checkout): `{"ride_id":"uuid","amount":500,"method":"cash"|"card"}`. Stub: immediately marked completed.
- **GET /api/v1/payments/ride/:rideId** — get payment by ride
- **GET /api/v1/payments/:id** — get payment by id
- **POST /api/v1/payments/:id/confirm** — stub confirm (e.g. cash on delivery)

## Env

- `PORT` (default 8084)
- `PG_DSN` (same as Auth)
- `JWT_SECRET` (must match Auth)
