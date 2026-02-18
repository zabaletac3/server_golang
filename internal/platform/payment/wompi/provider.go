package wompi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eren_dev/go_server/internal/platform/payment"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrInvalidPayload   = errors.New("invalid webhook payload")
	ErrWompiAPI         = errors.New("wompi api error")
	ErrNoWebhookSecret  = errors.New("webhook secret not configured")
)

const (
	sandboxURL    = "https://sandbox.wompi.co/v1"
	productionURL = "https://production.wompi.co/v1"
	checkoutURL   = "https://checkout.wompi.co/l"
)

// WompiProvider implementación del proveedor Wompi
type WompiProvider struct {
	publicKey     string
	privateKey    string
	webhookSecret string
	redirectURL   string
	baseURL       string
	httpClient    *http.Client
}

// NewWompiProvider crea una nueva instancia del proveedor Wompi.
// Detecta automáticamente sandbox vs producción según el prefijo de la key.
func NewWompiProvider(publicKey, privateKey, webhookSecret, redirectURL string) *WompiProvider {
	baseURL := productionURL
	if strings.HasPrefix(publicKey, "pub_test_") || strings.HasPrefix(privateKey, "prv_test_") {
		baseURL = sandboxURL
	}

	return &WompiProvider{
		publicKey:     publicKey,
		privateKey:    privateKey,
		webhookSecret: webhookSecret,
		redirectURL:   redirectURL,
		baseURL:       baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType retorna el tipo de proveedor
func (w *WompiProvider) GetProviderType() payment.ProviderType {
	return payment.ProviderWompi
}

// --- Payment Links API ---

type paymentLinkRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	SingleUse       bool   `json:"single_use"`
	CollectShipping bool   `json:"collect_shipping"`
	Currency        string `json:"currency"`
	AmountInCents   int64  `json:"amount_in_cents"`
	RedirectURL     string `json:"redirect_url,omitempty"`
	Sku             string `json:"sku,omitempty"`
}

type paymentLinkResponse struct {
	Data struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Active    bool   `json:"active"`
		Currency  string `json:"currency"`
		Amount    int64  `json:"amount_in_cents"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

// CreateSubscription crea un Payment Link en Wompi para que el usuario pague
func (w *WompiProvider) CreateSubscription(ctx context.Context, req *payment.SubscriptionRequest) (*payment.SubscriptionResponse, error) {
	redirectURL := req.RedirectURL
	if redirectURL == "" {
		redirectURL = w.redirectURL
	}

	linkReq := paymentLinkRequest{
		Name:            fmt.Sprintf("Suscripción %s", req.PlanName),
		Description:     fmt.Sprintf("Plan %s - %s | Tenant %s", req.PlanName, req.BillingPeriod, req.TenantID),
		SingleUse:       true,
		CollectShipping: false,
		Currency:        req.Currency,
		AmountInCents:   req.Amount,
		RedirectURL:     redirectURL,
		Sku:             fmt.Sprintf("plan_%s_%s", req.PlanID, req.TenantID),
	}

	jsonData, err := json.Marshal(linkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/payment_links", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+w.privateKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrWompiAPI, resp.StatusCode, string(body))
	}

	var linkResp paymentLinkResponse
	if err := json.Unmarshal(body, &linkResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Calcular próxima fecha de facturación
	var nextBilling *time.Time
	switch req.BillingPeriod {
	case "monthly":
		next := time.Now().AddDate(0, 1, 0)
		nextBilling = &next
	case "annual":
		next := time.Now().AddDate(1, 0, 0)
		nextBilling = &next
	}

	return &payment.SubscriptionResponse{
		SubscriptionID: linkResp.Data.ID,
		Status:         "PENDING",
		NextBillingAt:  nextBilling,
		Amount:         linkResp.Data.Amount,
		Currency:       linkResp.Data.Currency,
		PaymentLinkURL: fmt.Sprintf("%s/%s", checkoutURL, linkResp.Data.ID),
	}, nil
}

// CancelSubscription cancela una suscripción (se maneja en la DB, Wompi no tiene suscripciones nativas)
func (w *WompiProvider) CancelSubscription(ctx context.Context, subscriptionID string) error {
	return nil
}

// GetSubscription obtiene información de una transacción en Wompi
func (w *WompiProvider) GetSubscription(ctx context.Context, subscriptionID string) (*payment.SubscriptionResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", w.baseURL+"/transactions/"+subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+w.publicKey)

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrWompiAPI, resp.StatusCode, string(body))
	}

	var txResp struct {
		Data struct {
			ID            string `json:"id"`
			Status        string `json:"status"`
			AmountInCents int64  `json:"amount_in_cents"`
			Currency      string `json:"currency"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &txResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &payment.SubscriptionResponse{
		SubscriptionID: txResp.Data.ID,
		Status:         txResp.Data.Status,
		Amount:         txResp.Data.AmountInCents,
		Currency:       txResp.Data.Currency,
	}, nil
}

// ProcessWebhook procesa un webhook de Wompi
func (w *WompiProvider) ProcessWebhook(ctx context.Context, payload []byte, _ string) (*payment.WebhookEvent, error) {
	// Parsear payload para extraer signature.properties
	var webhookData struct {
		Event     string                 `json:"event"`
		Data      map[string]interface{} `json:"data"`
		SentAt    string                 `json:"sent_at"`
		Signature struct {
			Checksum   string   `json:"checksum"`
			Properties []string `json:"properties"`
		} `json:"signature"`
	}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, ErrInvalidPayload
	}

	// Validar firma
	if err := w.validateSignature(webhookData.Data, webhookData.Signature.Properties, webhookData.Signature.Checksum, webhookData.SentAt); err != nil {
		return nil, err
	}

	// Construir evento
	event := &payment.WebhookEvent{
		Provider:   payment.ProviderWompi,
		EventType:  webhookData.Event,
		RawPayload: payload,
		Metadata:   webhookData.Data,
	}

	// Extraer campos del data de la transacción
	if data := webhookData.Data; data != nil {
		if txData, ok := data["transaction"].(map[string]interface{}); ok {
			if id, ok := txData["id"].(string); ok {
				event.TransactionID = id
			}
			if status, ok := txData["status"].(string); ok {
				event.Status = status
			}
			if amount, ok := txData["amount_in_cents"].(float64); ok {
				event.Amount = int64(amount)
			}
			if currency, ok := txData["currency"].(string); ok {
				event.Currency = currency
			}
			if reference, ok := txData["reference"].(string); ok {
				event.SubscriptionID = reference
			}
		}
	}

	return event, nil
}

// validateSignature valida la firma del webhook según el algoritmo de Wompi:
// SHA256( concat(values from properties) + sent_at + events_secret )
func (w *WompiProvider) validateSignature(data map[string]interface{}, properties []string, checksum, sentAt string) error {
	if w.webhookSecret == "" {
		return ErrNoWebhookSecret
	}

	// Concatenar valores de las propiedades especificadas
	var concat string
	for _, prop := range properties {
		val := extractNestedValue(data, prop)
		concat += fmt.Sprintf("%v", val)
	}

	// Agregar timestamp y secret
	concat += sentAt
	concat += w.webhookSecret

	// Calcular SHA256
	hash := sha256.Sum256([]byte(concat))
	calculated := hex.EncodeToString(hash[:])

	if calculated != checksum {
		return ErrInvalidSignature
	}

	return nil
}

// extractNestedValue extrae un valor de un mapa usando dot notation (e.g., "transaction.id")
func extractNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = m[part]
		if !ok {
			return ""
		}
	}

	return current
}
