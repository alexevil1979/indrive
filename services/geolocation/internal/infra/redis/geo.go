// Package redis — Redis GEO: driver positions (GEORADIUS for nearest)
package redis

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/ridehail/geolocation/internal/domain"
)

type GeoStore struct {
	cli *redis.Client
	key string
}

func NewGeoStore(cli *redis.Client) *GeoStore {
	if cli == nil {
		return nil
	}
	return &GeoStore{cli: cli, key: GeoKeyDrivers}
}

// Update driver position (GEOADD)
func (s *GeoStore) Set(ctx context.Context, driverID string, lat, lng float64) error {
	return s.cli.GeoAdd(ctx, s.key, &redis.GeoLocation{
		Name:      driverID,
		Longitude: lng,
		Latitude:  lat,
	}).Err()
}

// Nearest drivers (GEORADIUS with WITHDIST, LIMIT)
func (s *GeoStore) Nearest(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]domain.DriverLocation, error) {
	if limit <= 0 {
		limit = 10
	}
	q := &redis.GeoRadiusQuery{
		Radius:    radiusKm,
		Unit:      "km",
		WithDist:  true,
		WithCoord: true,
		Count:     limit,
		Sort:      "ASC",
	}
	results, err := s.cli.GeoRadius(ctx, s.key, lng, lat, q).Result()
	if err != nil {
		return nil, err
	}
	out := make([]domain.DriverLocation, 0, len(results))
	for _, r := range results {
		loc := domain.DriverLocation{
			DriverID: r.Name,
			Location: domain.Location{Lat: r.Latitude, Lng: r.Longitude},
			Distance: r.Dist,
		}
		out = append(out, loc)
	}
	return out, nil
}

// Remove driver (e.g. offline) — ZREM by member name (Redis GEO is sorted set)
func (s *GeoStore) Remove(ctx context.Context, driverID string) error {
	return s.cli.ZRem(ctx, s.key, driverID).Err()
}
