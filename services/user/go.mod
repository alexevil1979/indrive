module github.com/ridehail/user

go 1.23

require (
	github.com/alexevil1979/indrive/packages/otel-go v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.7.1
	github.com/labstack/echo/v4 v4.12.0
	github.com/minio/minio-go/v7 v7.0.80
	github.com/redis/go-redis/v9 v9.7.0
)

replace github.com/alexevil1979/indrive/packages/otel-go => ../../packages/otel-go
