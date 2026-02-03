// Package gateway — stub payment gateways (cash, card)
// Later: Tinkoff, Sber, YooMoney integrations
package gateway

import (
	"context"
	"fmt"
)

type CashGateway struct{}

func (g *CashGateway) Charge(ctx context.Context, amount float64, currency, rideID string) (externalID string, err error) {
	// Cash: no external call, just return stub id
	return fmt.Sprintf("cash_stub_%s", rideID), nil
}

type CardGateway struct{}

func (g *CardGateway) Charge(ctx context.Context, amount float64, currency, rideID string) (externalID string, err error) {
	// Card stub: simulate success (later: Tinkoff/Sber/YooMoney)
	return fmt.Sprintf("card_stub_%s", rideID), nil
}
