package aex

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
)

var fakePDF = []byte("%PDF-1.4\nfake guide content\n%%EOF\n")

// printEndpoint serves the auth endpoint and replies to every other path
// with the given handler, capturing the decoded print request payload.
func printEndpoint(t *testing.T, captured any, reply http.HandlerFunc) http.Handler {
	return countingAuthEndpoint(t, new(int), func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/envios/imprimir" {
			t.Errorf("path = %q, want /envios/imprimir", r.URL.Path)
		}
		if captured != nil {
			if err := jsonDecodeStrict(r, captured); err != nil {
				t.Errorf("decode request: %v", err)
			}
		}
		reply(w, r)
	})
}

func TestPrintGuidePDF(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, printEndpoint(t, &captured, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(fakePDF)
	}))

	pdf, err := client.PrintGuide(context.Background(), "A002866303", FormatGuideA4, true)
	if err != nil {
		t.Fatalf("PrintGuide: %v", err)
	}
	if !bytes.Equal(pdf, fakePDF) {
		t.Errorf("pdf = %q, want %q", pdf, fakePDF)
	}

	if captured["guia"] != "A002866303" {
		t.Errorf("guia = %v", captured["guia"])
	}
	if captured["formato"] != FormatGuideA4 {
		t.Errorf("formato = %v, want %q", captured["formato"], FormatGuideA4)
	}
	if captured["imprimir_partida"] != true {
		t.Errorf("imprimir_partida = %v, want true", captured["imprimir_partida"])
	}
	if captured["clave_publica"] != testPublicKey {
		t.Errorf("clave_publica = %v", captured["clave_publica"])
	}
}

func TestPrintGuideFormatOmitted(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, printEndpoint(t, &captured, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(fakePDF)
	}))

	if _, err := client.PrintGuide(context.Background(), "A002866303", "", false); err != nil {
		t.Fatalf("PrintGuide: %v", err)
	}
	if _, ok := captured["formato"]; ok {
		t.Error("formato should be omitted when empty (API uses its default format)")
	}
	if captured["imprimir_partida"] != false {
		t.Errorf("imprimir_partida = %v, want false sent explicitly", captured["imprimir_partida"])
	}
}

func TestPrintGuideJSONError(t *testing.T) {
	client, _ := newTestClient(t, printEndpoint(t, nil, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"4","mensaje":"guia inexistente"}`))
	}))

	pdf, err := client.PrintGuide(context.Background(), "A999999999", FormatLabel8x6, false)
	if pdf != nil {
		t.Errorf("pdf = %q, want nil", pdf)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Code != "4" || apiErr.Message != "guia inexistente" {
		t.Errorf("apiErr = %+v", apiErr)
	}
	if apiErr.Endpoint != printGuidePath {
		t.Errorf("Endpoint = %q, want %q", apiErr.Endpoint, printGuidePath)
	}
}

func TestPrintGuideRetriesWithStaleToken(t *testing.T) {
	endpointCalls := 0
	var endpointTokens []string

	client, _ := newTestClient(t, countingAuthEndpoint(t, new(int), func(w http.ResponseWriter, r *http.Request) {
		endpointCalls++
		var req printGuideParams
		if err := jsonDecodeStrict(r, &req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		endpointTokens = append(endpointTokens, req.AuthorizationCode)

		if endpointCalls == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"codigo":"8","mensaje":"codigo de autorizacion invalido"}`))
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(fakePDF)
	}))

	// Prime the cache so the first print call reuses a stale code.
	if _, err := client.Authenticate(context.Background()); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	pdf, err := client.PrintGuide(context.Background(), "A002866303", FormatLabel65x45, false)
	if err != nil {
		t.Fatalf("PrintGuide: %v", err)
	}
	if !bytes.Equal(pdf, fakePDF) {
		t.Errorf("pdf = %q, want %q", pdf, fakePDF)
	}
	if endpointCalls != 2 {
		t.Errorf("endpointCalls = %d, want 2 (retry after re-auth)", endpointCalls)
	}
	if len(endpointTokens) != 2 || endpointTokens[0] != "token-1" || endpointTokens[1] != "token-2" {
		t.Errorf("endpointTokens = %v, want [token-1 token-2]", endpointTokens)
	}
}

func TestPrintGuideValidation(t *testing.T) {
	client, _ := newTestClient(t, printEndpoint(t, nil, func(w http.ResponseWriter, r *http.Request) {
		t.Error("endpoint should not be called with an empty guide number")
	}))

	if _, err := client.PrintGuide(context.Background(), "", FormatGuideA4, false); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestIsJSONContentType(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"Application/JSON", true},
		{"application/vnd.api+json", true},
		{"application/pdf", false},
		{"", false},
		{"text/html", false},
	}

	for _, tt := range tests {
		if got := isJSONContentType(tt.contentType); got != tt.want {
			t.Errorf("isJSONContentType(%q) = %t, want %t", tt.contentType, got, tt.want)
		}
	}
}
