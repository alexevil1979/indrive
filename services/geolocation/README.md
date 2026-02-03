# Geolocation Service (Go)

Driver tracking (Redis GEO), nearest drivers search, WebSocket stub for real-time tracking.

## Run locally

1. Start Redis: `docker compose -f ../../infra/docker-compose.yml up -d redis`
2. `go mod tidy && go run .`
3. Update driver location: `POST http://localhost:8082/api/v1/drivers/:driver_id/location` — `{"lat":55.75,"lng":37.62}`
4. Nearest drivers: `GET http://localhost:8082/api/v1/drivers/nearest?lat=55.75&lng=37.62&radius_km=5&limit=10`
5. WebSocket stub: `GET http://localhost:8082/ws/tracking` (upgrade to WS; receives `{"type":"connected","service":"geolocation","message":"tracking stub"}`)

## Env

- `PORT` (default 8082)
- `REDIS_ADDR` (default localhost:6379)
