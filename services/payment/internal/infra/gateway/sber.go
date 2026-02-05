// Package gateway — Sberbank Acquiring / SberPay integration
// Docs: https://securepayments.sberbank.ru/wiki/doku.php/integration:api
package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ridehail/payment/internal/domain"
)

const (
	sberAPIURL     = "https://securepayments.sberbank.ru/payment/rest"
	sberTestAPIURL = "https://3dsec.sberbank.ru/payment/rest"
)

// SberConfig — Sberbank gateway configuration
type SberConfig struct {
	UserName  string // API username (or token)
	Password  string // API password
	Token     string // Alternative: API token
	TestMode  bool   // Use test API
	ReturnURL string // Default return URL
	FailURL   string // Default fail URL
}

// SberGateway — Sberbank Acquiring gateway
type SberGateway struct {
	userName  string
	password  string
	token     string
	apiURL    string
	returnURL string
	failURL   string
	client    *http.Client
}

// NewSberGateway creates Sberbank gateway
func NewSberGateway(cfg SberConfig) *SberGateway {
	apiURL := sberAPIURL
	if cfg.TestMode {
		apiURL = sberTestAPIURL
	}
	return &SberGateway{
		userName:  cfg.UserName,
		password:  cfg.Password,
		token:     cfg.Token,
		apiURL:    apiURL,
		returnURL: cfg.ReturnURL,
		failURL:   cfg.FailURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Provider returns provider name
func (g *SberGateway) Provider() string {
	return domain.ProviderSber
}

// sberRegisterResponse — register.do response
type sberRegisterResponse struct {
	OrderId   string `json:"orderId"`
	FormUrl   string `json:"formUrl"`
	ErrorCode string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// sberOrderStatus — getOrderStatusExtended response
type sberOrderStatus struct {
	OrderNumber      string `json:"orderNumber"`
	OrderStatus      int    `json:"orderStatus"` // 0-registered, 1-preauth, 2-paid, 3-cancelled, 4-refunded, 5-ACS, 6-declined
	ActionCode       int    `json:"actionCode"`
	ActionCodeDescription string `json:"actionCodeDescription"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
	ErrorCode        string `json:"errorCode"`
	ErrorMessage     string `json:"errorMessage"`
	CardAuthInfo     *sberCardInfo `json:"cardAuthInfo,omitempty"`
	BindingInfo      *sberBindingInfo `json:"bindingInfo,omitempty"`
	MerchantOrderParams []sberOrderParam `json:"merchantOrderParams,omitempty"`
}

type sberCardInfo struct {
	Pan        string `json:"pan"`
	Expiration string `json:"expiration"` // YYYYMM
	CardholderName string `json:"cardholderName"`
}

type sberBindingInfo struct {
	BindingId string `json:"bindingId"`
	ClientId  string `json:"clientId"`
}

type sberOrderParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CreatePayment initiates payment via Sberbank
func (g *SberGateway) CreatePayment(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	// Amount in kopecks
	amountKopecks := int64(input.Amount * 100)

	returnURL := input.ReturnURL
	if returnURL == "" {
		returnURL = g.returnURL
	}
	failURL := g.failURL
	if failURL == "" {
		failURL = returnURL
	}

	params := url.Values{}
	g.addAuth(params)
	params.Set("orderNumber", input.PaymentID)
	params.Set("amount", strconv.FormatInt(amountKopecks, 10))
	params.Set("currency", "643") // RUB
	params.Set("returnUrl", returnURL)
	params.Set("failUrl", failURL)
	params.Set("description", input.Description)

	// Save card for recurring
	if input.SaveCard && input.Metadata["user_id"] != "" {
		params.Set("clientId", input.Metadata["user_id"])
	}

	// Use saved card
	if input.TokenID != "" {
		return g.payWithBinding(ctx, input)
	}

	// Add metadata
	if input.Metadata != nil {
		jsonParams := make([]map[string]string, 0)
		for k, v := range input.Metadata {
			jsonParams = append(jsonParams, map[string]string{"name": k, "value": v})
		}
		jsonData, _ := json.Marshal(jsonParams)
		params.Set("jsonParams", string(jsonData))
	}

	// Send request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/register.do", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var regResp sberRegisterResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if regResp.ErrorCode != "" && regResp.ErrorCode != "0" {
		return nil, fmt.Errorf("%w: %s - %s", ErrPaymentRejected, regResp.ErrorCode, regResp.ErrorMessage)
	}

	return &CreatePaymentResult{
		ExternalID:       regResp.OrderId,
		ConfirmURL:       regResp.FormUrl,
		Status:           domain.PaymentStatusPending,
		RequiresRedirect: true,
	}, nil
}

// payWithBinding pays with saved card
func (g *SberGateway) payWithBinding(ctx context.Context, input CreatePaymentInput) (*CreatePaymentResult, error) {
	amountKopecks := int64(input.Amount * 100)

	// First register order
	params := url.Values{}
	g.addAuth(params)
	params.Set("orderNumber", input.PaymentID)
	params.Set("amount", strconv.FormatInt(amountKopecks, 10))
	params.Set("currency", "643")
	params.Set("returnUrl", g.returnURL)
	params.Set("clientId", input.Metadata["user_id"])

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/register.do", strings.NewReader(params.Encode()))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var regResp sberRegisterResponse
	json.Unmarshal(respBody, &regResp)

	if regResp.ErrorCode != "" && regResp.ErrorCode != "0" {
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, regResp.ErrorMessage)
	}

	// Pay with binding
	params = url.Values{}
	g.addAuth(params)
	params.Set("mdOrder", regResp.OrderId)
	params.Set("bindingId", input.TokenID)

	httpReq, _ = http.NewRequestWithContext(ctx, "POST", g.apiURL+"/paymentOrderBinding.do", strings.NewReader(params.Encode()))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ = io.ReadAll(resp.Body)
	var bindResp struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
		Redirect     string `json:"redirect"`
		Info         string `json:"info"`
	}
	json.Unmarshal(respBody, &bindResp)

	if bindResp.ErrorCode != "" && bindResp.ErrorCode != "0" {
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, bindResp.ErrorMessage)
	}

	// Check if redirect needed (3DS)
	if bindResp.Redirect != "" {
		return &CreatePaymentResult{
			ExternalID:       regResp.OrderId,
			ConfirmURL:       bindResp.Redirect,
			Status:           domain.PaymentStatusProcessing,
			RequiresRedirect: true,
		}, nil
	}

	return &CreatePaymentResult{
		ExternalID:       regResp.OrderId,
		Status:           domain.PaymentStatusCompleted,
		RequiresRedirect: false,
	}, nil
}

// GetPaymentStatus gets payment status
func (g *SberGateway) GetPaymentStatus(ctx context.Context, externalID string) (string, error) {
	params := url.Values{}
	g.addAuth(params)
	params.Set("orderId", externalID)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/getOrderStatusExtended.do", strings.NewReader(params.Encode()))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var status sberOrderStatus
	json.Unmarshal(respBody, &status)

	return mapSberStatus(status.OrderStatus), nil
}

// Refund processes refund
func (g *SberGateway) Refund(ctx context.Context, input RefundInput) (*RefundResult, error) {
	amountKopecks := int64(input.Amount * 100)

	params := url.Values{}
	g.addAuth(params)
	params.Set("orderId", input.ExternalID)
	params.Set("amount", strconv.FormatInt(amountKopecks, 10))

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/refund.do", strings.NewReader(params.Encode()))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var refundResp struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	json.Unmarshal(respBody, &refundResp)

	if refundResp.ErrorCode != "" && refundResp.ErrorCode != "0" {
		return nil, fmt.Errorf("%w: %s", ErrPaymentRejected, refundResp.ErrorMessage)
	}

	return &RefundResult{
		RefundID:   input.ExternalID + "-refund",
		ExternalID: input.ExternalID,
		Amount:     input.Amount,
		Status:     "succeeded",
	}, nil
}

// sberCallback — Sberbank callback notification
type sberCallback struct {
	MdOrder    string `json:"mdOrder"`
	OrderNumber string `json:"orderNumber"`
	Operation  string `json:"operation"` // approved, deposited, reversed, refunded
	Status     int    `json:"status"`
	Amount     int64  `json:"amount"`
	Checksum   string `json:"checksum"`
}

// ParseWebhook parses Sberbank callback
func (g *SberGateway) ParseWebhook(ctx context.Context, body []byte, signature string) (*domain.WebhookEvent, error) {
	var wh sberCallback
	if err := json.Unmarshal(body, &wh); err != nil {
		// Try form-encoded
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, fmt.Errorf("parse webhook: %w", err)
		}
		wh.MdOrder = values.Get("mdOrder")
		wh.OrderNumber = values.Get("orderNumber")
		wh.Operation = values.Get("operation")
		wh.Status, _ = strconv.Atoi(values.Get("status"))
		wh.Amount, _ = strconv.ParseInt(values.Get("amount"), 10, 64)
		wh.Checksum = values.Get("checksum")
	}

	// Verify checksum if password is set
	if g.password != "" && wh.Checksum != "" {
		data := fmt.Sprintf("%s;%s;%d", wh.MdOrder, wh.Operation, wh.Amount)
		hash := sha256.Sum256([]byte(data + g.password))
		expectedChecksum := hex.EncodeToString(hash[:])
		if !strings.EqualFold(wh.Checksum, expectedChecksum) {
			return nil, ErrInvalidWebhook
		}
	}

	eventType := "payment.unknown"
	switch wh.Operation {
	case "approved", "deposited":
		eventType = "payment.succeeded"
	case "reversed":
		eventType = "payment.cancelled"
	case "refunded":
		eventType = "refund.succeeded"
	case "declined":
		eventType = "payment.failed"
	}

	status := mapSberStatus(wh.Status)

	return &domain.WebhookEvent{
		Provider:   domain.ProviderSber,
		EventType:  eventType,
		PaymentID:  wh.OrderNumber,
		ExternalID: wh.MdOrder,
		Status:     status,
		Amount:     float64(wh.Amount) / 100,
		RawPayload: string(body),
	}, nil
}

// GetSavedCard returns saved card (binding) info
func (g *SberGateway) GetSavedCard(ctx context.Context, externalID string) (*CardInfo, error) {
	params := url.Values{}
	g.addAuth(params)
	params.Set("orderId", externalID)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.apiURL+"/getOrderStatusExtended.do", strings.NewReader(params.Encode()))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var status sberOrderStatus
	json.Unmarshal(respBody, &status)

	if status.BindingInfo == nil || status.CardAuthInfo == nil {
		return nil, nil
	}

	// Parse expiration YYYYMM
	expMonth := 0
	expYear := 0
	if len(status.CardAuthInfo.Expiration) == 6 {
		expYear, _ = strconv.Atoi(status.CardAuthInfo.Expiration[:4])
		expMonth, _ = strconv.Atoi(status.CardAuthInfo.Expiration[4:])
	}

	// Determine brand
	pan := status.CardAuthInfo.Pan
	brand := "unknown"
	if strings.HasPrefix(pan, "4") {
		brand = "visa"
	} else if strings.HasPrefix(pan, "5") {
		brand = "mastercard"
	} else if strings.HasPrefix(pan, "2") {
		brand = "mir"
	}

	return &CardInfo{
		TokenID:     status.BindingInfo.BindingId,
		Last4:       pan[len(pan)-4:],
		Brand:       brand,
		ExpiryMonth: expMonth,
		ExpiryYear:  expYear,
	}, nil
}

// addAuth adds authentication params
func (g *SberGateway) addAuth(params url.Values) {
	if g.token != "" {
		params.Set("token", g.token)
	} else {
		params.Set("userName", g.userName)
		params.Set("password", g.password)
	}
}

// mapSberStatus maps Sberbank status to domain status
func mapSberStatus(status int) string {
	switch status {
	case 0:
		return domain.PaymentStatusPending
	case 1:
		return domain.PaymentStatusProcessing // Pre-authorized
	case 2:
		return domain.PaymentStatusCompleted
	case 3:
		return domain.PaymentStatusCancelled
	case 4:
		return domain.PaymentStatusRefunded
	case 5:
		return domain.PaymentStatusProcessing // ACS authorization
	case 6:
		return domain.PaymentStatusFailed
	default:
		return domain.PaymentStatusPending
	}
}
