// Package kafka — event producer for ride events (2026)
// Topics: ride.requested, ride.bid.placed, ride.matched, ride.status.changed
package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
)

const (
	TopicRideRequested   = "ride.requested"
	TopicRideBidPlaced   = "ride.bid.placed"
	TopicRideMatched     = "ride.matched"
	TopicRideStatusChanged = "ride.status.changed"
)

type Producer struct {
	prod sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	prod, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &Producer{prod: prod}, nil
}

func (p *Producer) Close() error {
	return p.prod.Close()
}

func (p *Producer) SendRideRequested(ctx context.Context, rideID, passengerID string, payload interface{}) error {
	return p.sendJSON(ctx, TopicRideRequested, rideID, payload)
}

func (p *Producer) SendRideBidPlaced(ctx context.Context, rideID, bidID, driverID string, price float64) error {
	payload := map[string]interface{}{"ride_id": rideID, "bid_id": bidID, "driver_id": driverID, "price": price}
	return p.sendJSON(ctx, TopicRideBidPlaced, rideID, payload)
}

func (p *Producer) SendRideMatched(ctx context.Context, rideID, driverID string, price float64) error {
	payload := map[string]interface{}{"ride_id": rideID, "driver_id": driverID, "price": price}
	return p.sendJSON(ctx, TopicRideMatched, rideID, payload)
}

func (p *Producer) SendRideStatusChanged(ctx context.Context, rideID, status string) error {
	payload := map[string]interface{}{"ride_id": rideID, "status": status}
	return p.sendJSON(ctx, TopicRideStatusChanged, rideID, payload)
}

func (p *Producer) sendJSON(ctx context.Context, topic, key string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(body),
	}
	_, _, err = p.prod.SendMessage(msg)
	if err != nil {
		slog.Warn("kafka send failed", "topic", topic, "error", err)
		return err
	}
	return nil
}
