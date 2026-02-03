# Ride Service (Go)

Request, bidding, matching, status + Kafka events (ride.requested, ride.bid.placed, ride.matched, ride.status.changed).

## Run locally

1. Start infra: `docker compose -f ../../infra/docker-compose.yml up -d postgres kafka`
2. Run Auth first (users + migrations for users/profiles)
3. `go mod tidy && go run .`
4. Get JWT from Auth (register/login). All ride endpoints require `Authorization: Bearer <token>`.
5. **Create ride** (passenger): `POST /api/v1/rides` — `{"from":{"lat":55.75,"lng":37.62,"address":"..."},"to":{"lat":55.76,"lng":37.63}}`
6. **Place bid** (driver): `POST /api/v1/rides/:id/bids` — `{"price":500}`
7. **List bids**: `GET /api/v1/rides/:id/bids`
8. **Accept bid** (passenger): `POST /api/v1/rides/:id/accept` — `{"bid_id":"..."}`
9. **Update status** (in_progress, completed, cancelled): `PATCH /api/v1/rides/:id/status` — `{"status":"in_progress"}`
10. **List my rides**: `GET /api/v1/rides?limit=20`
11. **List available rides** (driver only): `GET /api/v1/rides/available?limit=50` — rides in requested/bidding for drivers to bid
12. **List all rides** (admin only): `GET /api/v1/admin/rides?limit=100` — for admin panel dashboard/monitoring

## Env

- `PORT` (default 8083)
- `PG_DSN` (same as Auth)
- `KAFKA_BROKERS` (optional; empty = noop producer)
- `JWT_SECRET` (must match Auth)
