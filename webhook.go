package aex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WebhookEvent is an event notification that AEX sends to the client's
// configured webhook URL whenever a relevant event occurs in guide
// management.
type WebhookEvent struct {
	// GuideNumber is the transport guide number (series + number).
	GuideNumber string `json:"guia"`
	// Date is the event date and time.
	Date DateTime `json:"fecha"`
	// StatusCode is the event status code; empty when the status does not
	// change.
	StatusCode string `json:"codigo_estado"`
	// Status is the status description; empty when the status does not
	// change.
	Status string `json:"estado"`
	// EventTypeCode is the event type code; null or empty when the event
	// is of final status.
	EventTypeCode string `json:"codigo_tipo_evento"`
	// EventType is the event type description; null or empty when the
	// event is of final status.
	EventType string `json:"tipo_evento"`
	// Observation holds additional details; null or empty when absent.
	Observation string `json:"observacion"`
	// ClientOperationCode is the operation code reported when creating or
	// confirming the service; empty or null when the guide has no
	// associated operation code.
	ClientOperationCode string `json:"codigo_operacion_cliente"`
}

// ParseWebhook decodes a WebhookEvent from the body of an HTTP request
// sent by AEX.
func ParseWebhook(r *http.Request) (WebhookEvent, error) {
	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		return event, fmt.Errorf("aex: webhook: decode payload: %w", err)
	}
	return event, nil
}

// WebhookHandler processes a decoded webhook event. Returning nil makes
// HandleWebhook acknowledge the notification; any error makes it reply
// with the failure shape, which causes AEX to retry the notification.
type WebhookHandler func(ctx context.Context, event WebhookEvent) error

// webhookConfig collects the options of HandleWebhook.
type webhookConfig struct {
	authHeader string
	authValue  string
}

// WebhookOption customizes the handler returned by HandleWebhook.
type WebhookOption func(*webhookConfig)

// WithWebhookAuth requires the given HTTP header to carry exactly the
// given value (e.g., "Authorization", "Bearer <token>"), matching the
// authentication data configured with AEX for the webhook.
func WithWebhookAuth(header, value string) WebhookOption {
	return func(cfg *webhookConfig) {
		cfg.authHeader = header
		cfg.authValue = value
	}
}

// HandleWebhook returns an http.HandlerFunc that receives AEX webhook
// notifications: it validates the request, decodes the event, passes it to
// handler, and replies with the JSON shape the API expects:
//
//   - HTTP 200 {"isSuccess":true} after a successful decode and handler
//     execution
//   - HTTP 400 {"isSuccess":false} for invalid payloads, failed
//     authentication, or handler errors; AEX retries such notifications up
//     to 4 times (immediately, after 15, 30, and 60 minutes)
func HandleWebhook(handler WebhookHandler, opts ...WebhookOption) http.HandlerFunc {
	var cfg webhookConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeWebhookResult(w, http.StatusMethodNotAllowed, false)
			return
		}
		if cfg.authHeader != "" && r.Header.Get(cfg.authHeader) != cfg.authValue {
			writeWebhookResult(w, http.StatusBadRequest, false)
			return
		}

		event, err := ParseWebhook(r)
		if err != nil {
			writeWebhookResult(w, http.StatusBadRequest, false)
			return
		}
		if event.GuideNumber == "" {
			writeWebhookResult(w, http.StatusBadRequest, false)
			return
		}
		if err := handler(r.Context(), event); err != nil {
			writeWebhookResult(w, http.StatusBadRequest, false)
			return
		}
		writeWebhookResult(w, http.StatusOK, true)
	}
}

// writeWebhookResult writes the acknowledgement shape expected by AEX.
func writeWebhookResult(w http.ResponseWriter, status int, success bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		IsSuccess bool `json:"isSuccess"`
	}{success})
}
