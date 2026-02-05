// Package gateway — Tinkoff Acquiring integration
// Docs: https://www.tinkoff.ru/kassa/develop/api/payments/
package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ridehail/payment/internal/domain"
)

const (
	tinkoffAPIURL     = "https://securepay.tinkoff.ru/v2"
	tinkoffTestAPIURL = "https://rest-api-test.tinkoff.ru/v2"
)

// TinkoffConfig — Tinkoff gateway configuration
type TinkoffConfig struct {
	TerminalKey string // Terminal ID
	Password    string // Terminal password (for signature)
	TestMode    bool   // Use test API
	NotifyURL   string // Webhook URL
}

// TinkoffGateway — Tinkoff Acquiring gateway
type TinkoffGateway struct {
	terminalKey string
	password    string
	apiURL      string
	notifyURL   string
	client      *http.Client
}

// NewTinkoffGateway creates Tinkoff gateway
func NewTinkoffGateway(cfg TinkoffConfig) *TinkoffGateway {
	apiURL := tinkoffAPIURL
	if cfg.TestMode {
		apiURL = tinkoffTestAPIURL
	}
	return &TinkoffGateway{
		terminalKey: cfg.TerminalKey,
		password:    cfg.Password,
		apiURL:      apiURL,
		notifyURL:   cfg.NotifyURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Provider returns provider name
func (g *TinkoffGateway) Provider() string {
	return domain.ProviderTinkoff
}

// tinkoffInitRequest — Init payment request
type tinkoffInitRequest struct {
	TerminalKey string `json:"TerminalKey"`
	Amount      int64  `json:"Amount"` // In kopecks
	OrderId     string `json:"OrderId"`
	Description string `json:"Description,omitempty"`
	NotificationURL string `json:"NotificationURL,omitempty"`
	SuccessURL  string `json:"SuccessURL,omitempty"`
	FailURL     string `json:"FailURL,omitempty"`
	Recurrent   string `json:"Recurrent,omitempty"` // Y for save card
	CustomerKey string `json:"CustomerKey,omitempty"`
	Token       string `json:"Token"` // Signature
	DATA        map[string]string `json:"DATA,omitempty"`
	Receipt     *tinkoffReceipt `json:"Receipt,omitempty"`
}

// tinkoffReceipt — Receipt for fiscalization (54-FZ)
type tinkoffReceipt struct {
	Email    string          `json:"Email,omitempty"`
	Phone    string          `json:"Phone,omitempty"`
	Taxation string          `json:"Taxation"` // osn, usn_income, usn_income_outcome, patent, envd, esn
	Items    []tinkoffItem   `json:"Items"`
}

// tinkoffItem — Receipt item
type tinkoffItem struct {
	Name     string  `json:"Name"`
	Price    int64   `json:"Price"` // In kopecks
	Quantity float64 `json:"Quantity"`
	Amount   int64   `json:"Amount"` // Price * Quantity
	Tax      string  `json:"Tax"`    // none, vat0, vat10, vat20
}

// tinkoffInitResponse — Init payment response
type tinkoffInitResponse struct {
	Success    bool   `json:"Success"`
	ErrorCode  string `json:"ErrorCode"`
	Message    string `json:"Message"`
	TerminalKey string `json:"TerminalKey"`
	Status     string `json:"Status"`
	PaymentId  string `json:"PaymentId"`
	OrderId    string `json:"OrderId"`
	Amount     int64  `json:"Amount"`
	PaymentURL string `json:"PaymentURL"`
}

// CreatePayment initiates payment via Tinkoff
func (g *TinkoffGateway) CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	// Amount in kopecks
	amountKopecks := int64(input.Amount * 100)

	req := tinkoffInitRequest{
		TerminalKey:     g.terminalKey,
		Amount:          amountKopecks,
		OrderId:         input.PaymentID,
		Description:     input.Description,
		NotificationURL: g.notifyURL,
		SuccessURL:      input.ReturnURL,
		FailURL:         input.ReturnURL,
		DATA:            input.Metadata,
	}

	// Save card for recurring
	if input.SaveCard {
		req.Recurrent = "Y"
		req.CustomerKey = input.Metadata["user_id"]
	}

	// Use saved card
	if input.TokenID != "" {
		// For recurring payment, use ChargeWithRebillId instead
		return g.chargeRecurring(ctx, input)
	}

	// Add receipt for 54-FZ
	if input.UserEmail != "" || input.UserPhone != "" {
		req.Receipt = &tinkoffReceipt{
			Email:    input.UserEmail,
			Phone:    input.UserPhone,
			Taxation: "usn_income",
			Items: []tinkoffItem{
				{
					Name:     input.Description,
					Price:    amountKopecks,
					Quantity: 1,
					Amount:   amountKopecks,
					Tax:      "none",
				},
			},
		}
	}

	// Sign request
	req.Token = g.signRequest(map[string]interface{}{
		"TerminalKey": req.TerminalKey,
		"Amount":      req.Amount,
		"OrderId":     req.OrderId,
		"Description": req.Description,
	})

	// Send request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/Init", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var initResp tinkoffInitResponse
	if err := json.Unmarshal(respBody, &initResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if !initResp.Success {
		return nil, fmt.Errorf("%w: %s - %s", ErrPaymentRejected, initResp.ErrorCode, initResp.Message)
	}

	return &CreatePaymentResult{
		ExternalID:       initResp.PaymentId,
		ConfirmURL:       initResp.PaymentURL,
		Status:           mapTinkoffStatus(initResp.Status),
		RequiresRedirect: initResp.PaymentURL != "",
	}, nil
}

// chargeRecurring charges saved card
func (g *TinkoffGateway) chargeRecurring(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	// First init payment
	initReq := map[string]interface{}{
		"TerminalKey": g.terminalKey,
		"Amount":      int64(input.Amount * 100),
		"OrderId":     input.PaymentID,
		"Description": input.Description,
		"CustomerKey": input.Metadata["user_id"],
	}
	initReq["Token"] = g.signRequest(initReq)

	body, _ := json.Marshal(initReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/Init", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var initResp tinkoffInitResponse
	json.Unmarshal(respBody, &initResp)

	if !initResp.Success {
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, initResp.Message)
	}

	// Charge with saved card
	chargeReq := map[string]interface{}{
		"TerminalKey": g.terminalKey,
		"PaymentId":   initResp.PaymentId,
		"RebillId":    input.TokenID,
	}
	chargeReq["Token"] = g.signRequest(chargeReq)

	body, _ = json.Marshal(chargeReq)
	httpReq, _ = http.NewRequestWithContext(ctx, "POST", g.apiURL+"/Charge", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err = g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ = io.ReadAll(resp.Body)
	var chargeResp struct {
		Success   bool   `json:"Success"`
		Status    string `json:"Status"`
		PaymentId string `json:"PaymentId"`
		ErrorCode string `json:"ErrorCode"`
		Message   string `json:"Message"`
	}
	json.Unmarshal(respBody, &chargeResp)

	if !chargeResp.Success {
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, chargeResp.Message)
	}

	return &CreatePaymentResult{
		ExternalID:       chargeResp.PaymentId,
		Status:           mapTinkoffStatus(chargeResp.Status),
		RequiresRedirect: false,
	}, nil
}

// GetPaymentStatus gets payment status
func (g *TinkoffGateway) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	req := map[string]interface{}{
		"TerminalKey": g.terminalKey,
		"PaymentId":   externalID,
	}
	req["Token"] = g.signRequest(req)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/GetState", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stateResp struct {
		Success bool   `json:"Success"`
		Status  string `json:"Status"`
	}
	json.Unmarshal(respBody, &stateResp)

	return mapTinkoffStatus(stateResp.Status), nil
}

// Refund processes refund
func (g *TinkoffGateway) Refund(ctx context.Context, input RefundInput) (*RefundResult, error) {
	amountKopecks := int64(input.Amount * 100)

	req := map[string]interface{}{
		"TerminalKey": g.terminalKey,
		"PaymentId":   input.ExternalID,
		"Amount":      amountKopecks,
	}
	req["Token"] = g.signRequest(req)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/Cancel", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var cancelResp struct {
		Success     bool   `json:"Success"`
		Status      string `json:"Status"`
		PaymentId   string `json:"PaymentId"`
		OriginalAmount int64 `json:"OriginalAmount"`
		NewAmount   int64  `json:"NewAmount"`
		ErrorCode   string `json:"ErrorCode"`
		Message     string `json:"Message"`
	}
	json.Unmarshal(respBody, &cancelResp)

	if !cancelResp.Success {
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, cancelResp.Message)
	}

	return &RefundResult{
		RefundID:   cancelResp.PaymentId,
		ExternalID: input.ExternalID,
		Amount:     float64(cancelResp.OriginalAmount-cancelResp.NewAmount) / 100,
		Status:     mapTinkoffStatus(cancelResp.Status),
	}, nil
}

// tinkoffWebhook — Tinkoff notification
type tinkoffWebhook struct {
	TerminalKey string `json:"TerminalKey"`
	OrderId     string `json:"OrderId"`
	Success     bool   `json:"Success"`
	Status      string `json:"Status"`
	PaymentId   int64  `json:"PaymentId"`
	ErrorCode   string `json:"ErrorCode"`
	Amount      int64  `json:"Amount"`
	RebillId    int64  `json:"RebillId,omitempty"`
	CardId      int64  `json:"CardId,omitempty"`
	Pan         string `json:"Pan,omitempty"`
	ExpDate     string `json:"ExpDate,omitempty"` // MMYY
	Token       string `json:"Token"`
}

// ParseWebhook parses Tinkoff webhook
func (g *TinkoffGateway) ParseWebhook(ctx context.Context, body []byte, signature string) (*domain.WebhookEvent, error) {
	var wh tinkoffWebhook
	if err := json.Unmarshal(body, &wh); err != nil {
		return nil, fmt.Errorf("unmarshal webhook: %w", err)
	}

	// Verify signature
	params := make(map[string]interface{})
	json.Unmarshal(body, &params)
	delete(params, "Token")
	expectedToken := g.signRequest(params)

	if wh.Token != expectedToken {
		return nil, ErrInvalidWebhook
	}

	eventType := "payment.unknown"
	switch wh.Status {
	case "CONFIRMED", "AUTHORIZED":
		eventType = "payment.succeeded"
	case "REJECTED", "DEADLINE_EXPIRED":
		eventType = "payment.failed"
	case "REFUNDED", "PARTIAL_REFUNDED":
		eventType = "refund.succeeded"
	case "CANCELED":
		eventType = "payment.cancelled"
	}

	return &domain.WebhookEvent{
		Provider:   domain.ProviderTinkoff,
		EventType:  eventType,
		PaymentID:  wh.OrderId,
		ExternalID: strconv.FormatInt(wh.PaymentId, 10),
		Status:     mapTinkoffStatus(wh.Status),
		Amount:     float64(wh.Amount) / 100,
		RawPayload: string(body),
	}, nil
}

// GetSavedCard returns saved card info
func (g *TinkoffGateway) GetSavedCard(ctx context.Context, externalID string) (*CardInfo, error) {
	// Get card info from payment state
	req := map[string]interface{}{
		"TerminalKey": g.terminalKey,
		"PaymentId":   externalID,
	}
	req["Token"] = g.signRequest(req)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/GetState", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stateResp struct {
		RebillId int64  `json:"RebillId"`
		CardId   int64  `json:"CardId"`
		Pan      string `json:"Pan"`
		ExpDate  string `json:"ExpDate"` // MMYY
	}
	json.Unmarshal(respBody, &stateResp)

	if stateResp.RebillId == 0 {
		return nil, nil
	}

	var expMonth, expYear int
	if len(stateResp.ExpDate) == 4 {
		expMonth, _ = strconv.Atoi(stateResp.ExpDate[:2])
		expYear, _ = strconv.Atoi("20" + stateResp.ExpDate[2:])
	}

	// Determine brand from PAN prefix
	brand := "unknown"
	if strings.HasPrefix(stateResp.Pan, "4") {
		brand = "visa"
	} else if strings.HasPrefix(stateResp.Pan, "5") {
		brand = "mastercard"
	} else if strings.HasPrefix(stateResp.Pan, "2") {
		brand = "mir"
	}

	return &CardInfo{
		TokenID:     strconv.FormatInt(stateResp.RebillId, 10),
		Last4:       stateResp.Pan[len(stateResp.Pan)-4:],
		Brand:       brand,
		ExpiryMonth: expMonth,
		ExpiryYear:  expYear,
	}, nil
}

// signRequest creates Tinkoff signature
func (g *TinkoffGateway) signRequest(params map[string]interface{}) string {
	// Add password
	params["Password"] = g.password

	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Concatenate values
	var sb strings.Builder
	for _, k := range keys {
		v := params[k]
		switch val := v.(type) {
		case string:
			sb.WriteString(val)
		case int64:
			sb.WriteString(strconv.FormatInt(val, 10))
		case int:
			sb.WriteString(strconv.Itoa(val))
		case float64:
			sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
		case bool:
			if val {
				sb.WriteString("true")
			} else {
				sb.WriteString("false")
			}
		}
	}

	// SHA256
	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// mapTinkoffStatus maps Tinkoff status to domain status
func mapTinkoffStatus(status string) string {
	switch status {
	case "NEW", "FORM_SHOWED":
		return domain.PaymentStatusPending
	case "AUTHORIZING", "3DS_CHECKING", "3DS_CHECKED", "PREAUTHORIZING":
		return domain.PaymentStatusProcessing
	case "AUTHORIZED", "CONFIRMED":
		return domain.PaymentStatusCompleted
	case "REJECTED", "DEADLINE_EXPIRED":
		return domain.PaymentStatusFailed
	case "CANCELED":
		return domain.PaymentStatusCancelled
	case "REFUNDED", "PARTIAL_REFUNDED":
		return domain.PaymentStatusRefunded
	default:
		return domain.PaymentStatusPending
	}
}
