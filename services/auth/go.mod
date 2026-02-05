// RideHail auth service — Go 1.23+, DDD/Clean Architecture
// 2026: Echo (chosen over Fiber for ecosystem maturity), pgx, JWT
module github.com/ridehail/auth

go 1.23

require (
	github.com/alexevil1979/indrive/packages/otel-go v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.1
	github.com/labstack/echo/v4 v4.12.0
	github.com/redis/go-redis/v9 v9.7.0
	golang.org/x/crypto v0.28.0
)

replace github.com/alexevil1979/indrive/packages/otel-go => ../../packages/otel-go
