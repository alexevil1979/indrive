// Package gateway — YooMoney (ЮКасса) integration
// Docs: https://yookassa.ru/developers/api
package gateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ridehail/payment/internal/domain"
)

const yooMoneyAPIURL = "https://api.yookassa.ru/v3"

// YooMoneyConfig — YooMoney gateway configuration
type YooMoneyConfig struct {
	ShopID    string // Shop ID
	SecretKey string // Secret API key
	WebhookSecret string // Secret for webhook signature
	ReturnURL string // Default return URL
}

// YooMoneyGateway — YooMoney (ЮКасса) gateway
type YooMoneyGateway struct {
	shopID        string
	secretKey     string
	webhookSecret string
	returnURL     string
	client        *http.Client
}

// NewYooMoneyGateway creates YooMoney gateway
func NewYooMoneyGateway(cfg YooMoneyConfig) *YooMoneyGateway {
	return &YooMoneyGateway{
		shopID:        cfg.ShopID,
		secretKey:     cfg.SecretKey,
		webhookSecret: cfg.WebhookSecret,
		returnURL:     cfg.ReturnURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Provider returns provider name
func (g *YooMoneyGateway) Provider() string {
	return domain.ProviderYooMoney
}

// yooAmount — YooMoney amount
type yooAmount struct {
	Value    string `json:"value"`    // "100.00"
	Currency string `json:"currency"` // "RUB"
}

// yooConfirmation — Confirmation params
type yooConfirmation struct {
	Type      string `json:"type"`                 // redirect
	ReturnURL string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"` // Response only
}

// yooReceipt — Receipt for 54-FZ
type yooReceipt struct {
	Customer yooCustomer `json:"customer"`
	Items    []yooItem   `json:"items"`
}

type yooCustomer struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type yooItem struct {
	Description string    `json:"description"`
	Quantity    string    `json:"quantity"`
	Amount      yooAmount `json:"amount"`
	VatCode     int       `json:"vat_code"` // 1 = no VAT
}

// yooCreateRequest — Create payment request
type yooCreateRequest struct {
	Amount       yooAmount        `json:"amount"`
	Description  string           `json:"description,omitempty"`
	Confirmation *yooConfirmation `json:"confirmation,omitempty"`
	Capture      bool             `json:"capture"` // Auto-capture
	SavePaymentMethod bool        `json:"save_payment_method,omitempty"`
	PaymentMethodID string        `json:"payment_method_id,omitempty"` // For saved card
	Metadata     map[string]string `json:"metadata,omitempty"`
	Receipt      *yooReceipt      `json:"receipt,omitempty"`
}

// yooPayment — Payment object
type yooPayment struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"` // pending, waiting_for_capture, succeeded, canceled
	Amount        yooAmount       `json:"amount"`
	Description   string          `json:"description"`
	Confirmation  *yooConfirmation `json:"confirmation,omitempty"`
	PaymentMethod *yooPaymentMethod `json:"payment_method,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     string          `json:"created_at"`
	Paid          bool            `json:"paid"`
	Refundable    bool            `json:"refundable"`
}

type yooPaymentMethod struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Saved  bool   `json:"saved"`
	Title  string `json:"title,omitempty"`
	Card   *yooCard `json:"card,omitempty"`
}

type yooCard struct {
	First6      string `json:"first6"`
	Last4       string `json:"last4"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
	CardType    string `json:"card_type"` // Visa, MasterCard, Mir
}

// CreatePayment initiates payment via YooMoney
func (g *YooMoneyGateway) CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	returnURL := input.ReturnURL
	if returnURL == "" {
		returnURL = g.returnURL
	}

	req := yooCreateRequest{
		Amount: yooAmount{
			Value:    fmt.Sprintf("%.2f", input.Amount),
			Currency: input.Currency,
		},
		Description:       input.Description,
		Capture:           true,
		SavePaymentMethod: input.SaveCard,
		Metadata:          input.Metadata,
	}

	// Use saved card
	if input.TokenID != "" {
		req.PaymentMethodID = input.TokenID
	} else {
		// Redirect flow
		req.Confirmation = &yooConfirmation{
			Type:      "redirect",
			ReturnURL: returnURL,
		}
	}

	// Add payment_id to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]string)
	}
	req.Metadata["payment_id"] = input.PaymentID

	// Add receipt for 54-FZ
	if input.UserEmail != "" || input.UserPhone != "" {
		req.Receipt = &yooReceipt{
			Customer: yooCustomer{
				Email: input.UserEmail,
				Phone: input.UserPhone,
			},
			Items: []yooItem{
				{
					Description: input.Description,
					Quantity:    "1",
					Amount: yooAmount{
						Value:    fmt.Sprintf("%.2f", input.Amount),
						Currency: input.Currency,
					},
					VatCode: 1,
				},
			},
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", yooMoneyAPIURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", input.PaymentID) // Idempotency
	httpReq.SetBasicAuth(g.shopID, g.secretKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Type        string `json:"type"`
			ID          string `json:"id"`
			Code        string `json:"code"`
			Description string `json:"description"`
		}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%w: %s - %s", ErrPaymentRejected, errResp.Code, errResp.Description)
	}

	var payment yooPayment
	if err := json.Unmarshal(respBody, &payment); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	result := &CreatePaymentResult{
		ExternalID: payment.ID,
		Status:     mapYooMoneyStatus(payment.Status),
	}

	if payment.Confirmation != nil && payment.Confirmation.ConfirmationURL != "" {
		result.ConfirmURL = payment.Confirmation.ConfirmationURL
		result.RequiresRedirect = true
	}

	return result, nil
}

// GetPaymentStatus gets payment status
func (g *YooMoneyGateway) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", yooMoneyAPIURL+"/payments/"+externalID, nil)
	if err != nil {
		return "", err
	}
	httpReq.SetBasicAuth(g.shopID, g.secretKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var payment yooPayment
	json.Unmarshal(respBody, &payment)

	return mapYooMoneyStatus(payment.Status), nil
}

// Refund processes refund
func (g *YooMoneyGateway) Refund(ctx context.Context, input RefundInput) (*RefundResult, error) {
	req := map[string]interface{}{
		"payment_id": input.ExternalID,
		"amount": map[string]string{
			"value":    fmt.Sprintf("%.2f", input.Amount),
			"currency": "RUB",
		},
	}
	if input.Reason != "" {
		req["description"] = input.Reason
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", yooMoneyAPIURL+"/refunds", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", input.PaymentID+"-refund")
	httpReq.SetBasicAuth(g.shopID, g.secretKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var errResp struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, errResp.Description)
	}

	var refund struct {
		ID     string    `json:"id"`
		Status string    `json:"status"`
		Amount yooAmount `json:"amount"`
	}
	json.Unmarshal(respBody, &refund)

	return &RefundResult{
		RefundID:   refund.ID,
		ExternalID: input.ExternalID,
		Amount:     input.Amount,
		Status:     refund.Status,
	}, nil
}

// yooWebhook — YooMoney notification
type yooWebhook struct {
	Type   string `json:"type"` // notification
	Event  string `json:"event"` // payment.succeeded, payment.canceled, refund.succeeded
	Object yooPayment `json:"object"`
}

// ParseWebhook parses YooMoney webhook
func (g *YooMoneyGateway) ParseWebhook(ctx context.Context, body []byte, signature string) (*domain.WebhookEvent, error) {
	// Verify signature if webhook secret is set
	if g.webhookSecret != "" && signature != "" {
		mac := hmac.New(sha256.New, []byte(g.webhookSecret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if signature != expectedSig {
			return nil, ErrInvalidWebhook
		}
	}

	var wh yooWebhook
	if err := json.Unmarshal(body, &wh); err != nil {
		return nil, fmt.Errorf("unmarshal webhook: %w", err)
	}

	// Get our payment_id from metadata
	paymentID := ""
	if wh.Object.Metadata != nil {
		paymentID = wh.Object.Metadata["payment_id"]
	}

	return &domain.WebhookEvent{
		Provider:   domain.ProviderYooMoney,
		EventType:  wh.Event,
		PaymentID:  paymentID,
		ExternalID: wh.Object.ID,
		Status:     mapYooMoneyStatus(wh.Object.Status),
		Amount:     parseAmount(wh.Object.Amount.Value),
		RawPayload: string(body),
	}, nil
}

// GetSavedCard returns saved card info
func (g *YooMoneyGateway) GetSavedCard(ctx context.Context, externalID string) (*CardInfo, error) {
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", yooMoneyAPIURL+"/payments/"+externalID, nil)
	httpReq.SetBasicAuth(g.shopID, g.secretKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var payment yooPayment
	json.Unmarshal(respBody, &payment)

	if payment.PaymentMethod == nil || !payment.PaymentMethod.Saved || payment.PaymentMethod.Card == nil {
		return nil, nil
	}

	card := payment.PaymentMethod.Card
	expMonth := 0
	expYear := 0
	fmt.Sscanf(card.ExpiryMonth, "%d", &expMonth)
	fmt.Sscanf(card.ExpiryYear, "%d", &expYear)

	return &CardInfo{
		TokenID:     payment.PaymentMethod.ID,
		Last4:       card.Last4,
		Brand:       mapCardType(card.CardType),
		ExpiryMonth: expMonth,
		ExpiryYear:  expYear,
	}, nil
}

// mapYooMoneyStatus maps YooMoney status to domain status
func mapYooMoneyStatus(status string) string {
	switch status {
	case "pending":
		return domain.PaymentStatusPending
	case "waiting_for_capture":
		return domain.PaymentStatusProcessing
	case "succeeded":
		return domain.PaymentStatusCompleted
	case "canceled":
		return domain.PaymentStatusCancelled
	default:
		return domain.PaymentStatusPending
	}
}

// mapCardType maps YooMoney card type
func mapCardType(cardType string) string {
	switch cardType {
	case "Visa":
		return "visa"
	case "MasterCard":
		return "mastercard"
	case "Mir":
		return "mir"
	default:
		return "unknown"
	}
}

// parseAmount parses string amount to float64
func parseAmount(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
