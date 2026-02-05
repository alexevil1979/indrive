module github.com/ridehail/geolocation

go 1.23

require (
	github.com/alexevil1979/indrive/packages/otel-go v0.0.0
	github.com/gorilla/websocket v1.5.3
	github.com/labstack/echo/v4 v4.12.0
	github.com/redis/go-redis/v9 v9.7.0
)

replace github.com/alexevil1979/indrive/packages/otel-go => ../../packages/otel-go
