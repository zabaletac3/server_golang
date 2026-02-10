package wompi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eren_dev/go_server/internal/platform/payment"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrInvalidPayload   = errors.New("invalid webhook payload")
	ErrWompiAPI         = errors.New("wompi api error")
)

// WompiProvider implementación del proveedor Wompi
type WompiProvider struct {
	publicKey     string
	privateKey    string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

// NewWompiProvider crea una nueva instancia del proveedor Wompi
func NewWompiProvider(publicKey, privateKey, webhookSecret string) *WompiProvider {
	return &WompiProvider{
		publicKey:     publicKey,
		privateKey:    privateKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://production.wompi.co/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType retorna el tipo de proveedor
func (w *WompiProvider) GetProviderType() payment.ProviderType {
	return payment.ProviderWompi
}

// WompiTransactionRequest estructura para crear transacción
type WompiTransactionRequest struct {
	AmountInCents      int64                  `json:"amount_in_cents"`
	Currency           string                 `json:"currency"`
	CustomerEmail      string                 `json:"customer_email"`
	PaymentMethod      map[string]interface{} `json:"payment_method"`
	Reference          string                 `json:"reference"`
	CustomerData       map[string]string      `json:"customer_data,omitempty"`
	ShippingAddress    map[string]string      `json:"shipping_address,omitempty"`
	RedirectURL        string                 `json:"redirect_url,omitempty"`
	PaymentSourceID    int                    `json:"payment_source_id,omitempty"`
}

// WompiTransactionResponse estructura de respuesta de Wompi
type WompiTransactionResponse struct {
	Data struct {
		ID                string `json:"id"`
		Status            string `json:"status"`
		AmountInCents     int64  `json:"amount_in_cents"`
		Currency          string `json:"currency"`
		CustomerEmail     string `json:"customer_email"`
		Reference         string `json:"reference"`
		PaymentMethodType string `json:"payment_method_type"`
		PaymentLinkID     string `json:"payment_link_id,omitempty"`
	} `json:"data"`
	Meta struct {
		Trace string `json:"trace"`
	} `json:"meta"`
}

// CreateSubscription crea una nueva suscripción en Wompi
func (w *WompiProvider) CreateSubscription(ctx context.Context, req *payment.SubscriptionRequest) (*payment.SubscriptionResponse, error) {
	// Wompi no tiene suscripciones nativas, creamos una transacción recurrente
	// o un payment link para que el cliente pague
	
	transactionReq := WompiTransactionRequest{
		AmountInCents: req.Amount,
		Currency:      req.Currency,
		CustomerEmail: req.CustomerEmail,
		Reference:     fmt.Sprintf("subscription_%s_%d", req.TenantID, time.Now().Unix()),
		CustomerData: map[string]string{
			"full_name":    req.CustomerName,
			"phone_number": "",
		},
	}

	// Si tenemos payment_source_id (tarjeta tokenizada), usarlo
	if req.PaymentMethod != "" {
		transactionReq.PaymentMethod = map[string]interface{}{
			"type":         "CARD",
			"installments": 1,
		}
	}

	jsonData, err := json.Marshal(transactionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/transactions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Autenticación con Bearer token (private key en base64)
	auth := base64.StdEncoding.EncodeToString([]byte(w.privateKey + ":"))
	httpReq.Header.Set("Authorization", "Bearer "+auth)
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

	var wompiResp WompiTransactionResponse
	if err := json.Unmarshal(body, &wompiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Calcular próxima fecha de facturación según el período
	var nextBilling *time.Time
	if req.BillingPeriod == "monthly" {
		next := time.Now().AddDate(0, 1, 0)
		nextBilling = &next
	} else if req.BillingPeriod == "annual" {
		next := time.Now().AddDate(1, 0, 0)
		nextBilling = &next
	}

	return &payment.SubscriptionResponse{
		SubscriptionID: wompiResp.Data.ID,
		Status:         wompiResp.Data.Status,
		NextBillingAt:  nextBilling,
		Amount:         wompiResp.Data.AmountInCents,
		Currency:       wompiResp.Data.Currency,
	}, nil
}

// CancelSubscription cancela una suscripción en Wompi
func (w *WompiProvider) CancelSubscription(ctx context.Context, subscriptionID string) error {
	// Wompi no tiene endpoint de cancelación de suscripción
	// En un escenario real, deberías:
	// 1. Marcar en tu DB que la suscripción está cancelada
	// 2. No procesar más cargos recurrentes
	// 3. Opcionalmente, hacer un refund de la última transacción si aplica
	
	// Por ahora, solo retornamos nil (la cancelación se maneja en tu DB)
	return nil
}

// GetSubscription obtiene información de una suscripción
func (w *WompiProvider) GetSubscription(ctx context.Context, subscriptionID string) (*payment.SubscriptionResponse, error) {
	// Consultar el estado de una transacción en Wompi
	httpReq, err := http.NewRequestWithContext(ctx, "GET", w.baseURL+"/transactions/"+subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Autenticación
	auth := base64.StdEncoding.EncodeToString([]byte(w.publicKey + ":"))
	httpReq.Header.Set("Authorization", "Bearer "+auth)

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

	var wompiResp WompiTransactionResponse
	if err := json.Unmarshal(body, &wompiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &payment.SubscriptionResponse{
		SubscriptionID: wompiResp.Data.ID,
		Status:         wompiResp.Data.Status,
		Amount:         wompiResp.Data.AmountInCents,
		Currency:       wompiResp.Data.Currency,
	}, nil
}

// ProcessWebhook procesa un webhook de Wompi
func (w *WompiProvider) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	// Validar firma del webhook
	if !w.validateSignature(payload, signature) {
		return nil, ErrInvalidSignature
	}

	// Parsear payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, ErrInvalidPayload
	}

	// Extraer datos del evento
	event := &payment.WebhookEvent{
		Provider:   payment.ProviderWompi,
		RawPayload: payload,
		Metadata:   webhookData,
	}

	// Extraer campos específicos de Wompi
	if eventType, ok := webhookData["event"].(string); ok {
		event.EventType = eventType
	}

	if data, ok := webhookData["data"].(map[string]interface{}); ok {
		if transactionID, ok := data["id"].(string); ok {
			event.TransactionID = transactionID
		}
		if status, ok := data["status"].(string); ok {
			event.Status = status
		}
		if amountInCents, ok := data["amount_in_cents"].(float64); ok {
			event.Amount = int64(amountInCents)
		}
		if currency, ok := data["currency"].(string); ok {
			event.Currency = currency
		}
	}

	return event, nil
}

// validateSignature valida la firma del webhook
func (w *WompiProvider) validateSignature(payload []byte, signature string) bool {
	if w.webhookSecret == "" {
		// Si no hay secret configurado, aceptar (solo para desarrollo)
		return true
	}

	// Calcular HMAC SHA256
	mac := hmac.New(sha256.New, []byte(w.webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
