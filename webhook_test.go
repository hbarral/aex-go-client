package aex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const webhookPayload = `{
	"guia": "A123456",
	"fecha": "2025-06-06 12:34:56",
	"codigo_estado": "E",
	"estado": "Entregado",
	"codigo_tipo_evento": "W5",
	"tipo_evento": "Entrega Realizada",
	"observacion": "Paquete entregado correctamente",
	"codigo_operacion_cliente": "PEDIDO-12345"
}`

func TestParseWebhook(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(webhookPayload)))

	event, err := ParseWebhook(req)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}

	if event.GuideNumber != "A123456" {
		t.Errorf("GuideNumber = %q", event.GuideNumber)
	}
	if want := time.Date(2025, time.June, 6, 12, 34, 56, 0, time.UTC); !event.Date.Time.Equal(want) {
		t.Errorf("Date = %v, want %v", event.Date.Time, want)
	}
	if event.StatusCode != "E" || event.Status != "Entregado" {
		t.Errorf("status = %q/%q", event.StatusCode, event.Status)
	}
	if event.EventTypeCode != "W5" || event.EventType != "Entrega Realizada" {
		t.Errorf("event type = %q/%q", event.EventTypeCode, event.EventType)
	}
	if event.Observation != "Paquete entregado correctamente" {
		t.Errorf("Observation = %q", event.Observation)
	}
	if event.ClientOperationCode != "PEDIDO-12345" {
		t.Errorf("ClientOperationCode = %q", event.ClientOperationCode)
	}
}

func TestParseWebhookFinalStatus(t *testing.T) {
	// Final-status events carry null/empty event-type fields.
	payload := `{
		"guia": "A123456",
		"fecha": "2025-06-06 12:34:56",
		"codigo_estado": "E",
		"estado": "Entregado",
		"codigo_tipo_evento": null,
		"tipo_evento": "",
		"observacion": null,
		"codigo_operacion_cliente": null
	}`

	event, err := ParseWebhook(httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(payload))))
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	if event.EventTypeCode != "" || event.EventType != "" || event.Observation != "" {
		t.Errorf("event = %+v, want empty final-status fields", event)
	}
	if event.ClientOperationCode != "" {
		t.Errorf("ClientOperationCode = %q, want empty for null", event.ClientOperationCode)
	}
}

func TestParseWebhookInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte("not json")))
	if _, err := ParseWebhook(req); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandleWebhook(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		payload       string
		handler       WebhookHandler
		wantStatus    int
		wantIsSuccess bool
	}{
		{
			name:          "success",
			method:        http.MethodPost,
			payload:       webhookPayload,
			handler:       func(ctx context.Context, event WebhookEvent) error { return nil },
			wantStatus:    http.StatusOK,
			wantIsSuccess: true,
		},
		{
			name:       "handler error",
			method:     http.MethodPost,
			payload:    webhookPayload,
			handler:    func(ctx context.Context, event WebhookEvent) error { return errors.New("boom") },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			method:     http.MethodPost,
			payload:    "not json",
			handler:    func(ctx context.Context, event WebhookEvent) error { return nil },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing guide number",
			method:     http.MethodPost,
			payload:    `{"fecha": "2025-06-06 12:34:56"}`,
			handler:    func(ctx context.Context, event WebhookEvent) error { return nil },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			payload:    webhookPayload,
			handler:    func(ctx context.Context, event WebhookEvent) error { return nil },
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/webhook", bytes.NewReader([]byte(tt.payload)))

			HandleWebhook(tt.handler)(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			var result struct {
				IsSuccess bool `json:"isSuccess"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
			}
			if result.IsSuccess != tt.wantIsSuccess {
				t.Errorf("isSuccess = %t, want %t", result.IsSuccess, tt.wantIsSuccess)
			}
			if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}

func TestHandleWebhookPassesEvent(t *testing.T) {
	var got WebhookEvent

	handler := HandleWebhook(func(ctx context.Context, event WebhookEvent) error {
		got = event
		return nil
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(webhookPayload)))
	handler(recorder, req)

	if got.GuideNumber != "A123456" {
		t.Errorf("GuideNumber = %q, want A123456", got.GuideNumber)
	}
	if got.ClientOperationCode != "PEDIDO-12345" {
		t.Errorf("ClientOperationCode = %q", got.ClientOperationCode)
	}
}

func TestHandleWebhookAuth(t *testing.T) {
	process := func(ctx context.Context, event WebhookEvent) error { return nil }
	handler := HandleWebhook(process, WithWebhookAuth("Authorization", "Bearer secret-token"))

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"valid token", "Bearer secret-token", http.StatusOK},
		{"invalid token", "Bearer wrong", http.StatusBadRequest},
		{"missing header", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(webhookPayload)))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			handler(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandleWebhookContextCancellation(t *testing.T) {
	handler := HandleWebhook(func(ctx context.Context, event WebhookEvent) error {
		return ctx.Err()
	})

	recorder := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(webhookPayload))).WithContext(ctx)
	handler(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
