package kafka

import "context"

// NoopProducer — when Kafka is not configured (e.g. local dev)
type NoopProducer struct{}

func (p *NoopProducer) SendRideRequested(ctx context.Context, rideID, passengerID string, payload interface{}) error {
	return nil
}

func (p *NoopProducer) SendRideBidPlaced(ctx context.Context, rideID, bidID, driverID string, price float64) error {
	return nil
}

func (p *NoopProducer) SendRideMatched(ctx context.Context, rideID, driverID string, price float64) error {
	return nil
}

func (p *NoopProducer) SendRideStatusChanged(ctx context.Context, rideID, status string) error {
	return nil
}
