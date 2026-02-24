package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrMissingSignature is returned when the signature header is empty
	ErrMissingSignature = errors.New("missing webhook signature")
	// ErrInvalidSignature is returned when the signature validation fails
	ErrInvalidSignature = errors.New("invalid webhook signature")
)

// SignatureValidator validates webhook signatures for different providers
type SignatureValidator struct {
	secrets map[string]string // provider -> secret
}

// NewSignatureValidator creates a new signature validator
func NewSignatureValidator() *SignatureValidator {
	return &SignatureValidator{
		secrets: make(map[string]string),
	}
}

// RegisterSecret registers a secret for a specific provider
func (v *SignatureValidator) RegisterSecret(provider, secret string) {
	v.secrets[strings.ToLower(provider)] = secret
}

// Validate validates a webhook signature
// provider: the payment provider name (wompi, stripe, etc.)
// signature: the signature from the request header
// payload: the raw request body
func (v *SignatureValidator) Validate(provider, signature, payload string) error {
	if signature == "" {
		return ErrMissingSignature
	}

	secret, ok := v.secrets[strings.ToLower(provider)]
	if !ok || secret == "" {
		// No secret configured for this provider, skip validation
		// This allows development/testing without secrets
		return nil
	}

	return validateHMACSignature(signature, payload, secret)
}

// validateHMACSignature validates HMAC-SHA256 signature
func validateHMACSignature(signature, payload, secret string) error {
	// Compute expected signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures using constant-time comparison
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("%w: expected %s, got %s", ErrInvalidSignature, expectedSignature, signature)
	}

	return nil
}

// ValidateWompi validates Wompi webhook signature
// Wompi uses X-Signature or X-Wompi-Signature header with HMAC-SHA256
func (v *SignatureValidator) ValidateWompi(signature, payload string) error {
	return v.Validate("wompi", signature, payload)
}

// ValidateStripe validates Stripe webhook signature
// Stripe uses Stripe-Signature header with different format (t=timestamp, v1=signature)
func (v *SignatureValidator) ValidateStripe(signature, payload string) error {
	if signature == "" {
		return ErrMissingSignature
	}

	secret, ok := v.secrets["stripe"]
	if !ok || secret == "" {
		return nil // Skip validation if no secret configured
	}

	// Stripe signature format: t=timestamp,v1=signature,v0=signature
	// Parse the signature
	var timestamp, sigV1 string
	parts := strings.Split(signature, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			sigV1 = kv[1]
		}
	}

	if timestamp == "" || sigV1 == "" {
		return fmt.Errorf("%w: invalid stripe signature format", ErrInvalidSignature)
	}

	// Build signed payload: timestamp.payload
	signedPayload := fmt.Sprintf("%s.%s", timestamp, payload)

	return validateHMACSignature(sigV1, signedPayload, secret)
}
