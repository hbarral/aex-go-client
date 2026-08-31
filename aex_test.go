package aex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{"missing public key", Config{PrivateKey: "priv", SessionCode: "session"}},
		{"missing private key", Config{PublicKey: "pub", SessionCode: "session"}},
		{"missing session code", Config{PublicKey: "pub", PrivateKey: "priv"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.config); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestNewBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		sandbox bool
		want    string
	}{
		{"production", false, ProductionBaseURL},
		{"sandbox", true, SandboxBaseURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(Config{PublicKey: "pub", PrivateKey: "priv", SessionCode: "session", Sandbox: tt.sandbox})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if client.baseURL != tt.want {
				t.Errorf("baseURL = %q, want %q", client.baseURL, tt.want)
			}
		})
	}

	t.Run("with base url override", func(t *testing.T) {
		client, err := New(Config{PublicKey: "pub", PrivateKey: "priv", SessionCode: "session"}, WithBaseURL("http://localhost:1/"))
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if client.baseURL != "http://localhost:1/" {
			t.Errorf("baseURL = %q, want override", client.baseURL)
		}
	})
}

type envelopeResponse struct {
	baseResponse
	Datos json.RawMessage `json:"datos"`
}

func TestDoRequest(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantErrIs  error
		wantErrMsg string
	}{
		{
			name: "success decodes response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","datos":[]}`))
			},
		},
		{
			name: "non-2xx status returns HTTPError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErrIs: &HTTPError{},
		},
		{
			name: "invalid json returns error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`not json`))
			},
			wantErrMsg: "decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := New(Config{PublicKey: "pub", PrivateKey: "priv", SessionCode: "session"}, WithBaseURL(server.URL+"/"))
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			var out envelopeResponse
			err = client.doRequest(context.Background(), "/envios/ciudades", map[string]string{}, &out)
			if tt.wantErrIs != nil || tt.wantErrMsg != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var httpErr *HTTPError
				if tt.wantErrIs != nil && !errors.As(err, &httpErr) {
					t.Errorf("error = %v, want %T", err, tt.wantErrIs)
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("doRequest: %v", err)
			}
			if out.Codigo != "0" {
				t.Errorf("Codigo = %q, want 0", out.Codigo)
			}
		})
	}
}

func TestDoRequestNetworkError(t *testing.T) {
	client, err := New(Config{PublicKey: "pub", PrivateKey: "priv", SessionCode: "session"}, WithBaseURL("http://127.0.0.1:0/"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	err = client.doRequest(context.Background(), "/envios/ciudades", map[string]string{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDoRequestContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client, err := New(Config{PublicKey: "pub", PrivateKey: "priv", SessionCode: "session"}, WithBaseURL(server.URL+"/"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.doRequest(ctx, "/envios/ciudades", map[string]string{}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name     string
		codigo   ResultCode
		mensaje  string
		wantErr  bool
		wantCode string
	}{
		{name: "zero code is success", codigo: "0", mensaje: "OK"},
		{name: "non-zero code is error", codigo: "12", mensaje: "invalid credentials", wantErr: true, wantCode: "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkResponse("/test", baseResponse{Codigo: tt.codigo, Mensaje: tt.mensaje})
			if tt.wantErr {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("error = %v, want *APIError", err)
				}
				if apiErr.Code != tt.wantCode {
					t.Errorf("Code = %q, want %q", apiErr.Code, tt.wantCode)
				}
				if apiErr.Message != tt.mensaje {
					t.Errorf("Message = %q, want %q", apiErr.Message, tt.mensaje)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
