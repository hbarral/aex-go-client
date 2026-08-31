package aex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

const (
	testPublicKey   = "pub-key"
	testPrivateKey  = "abc"
	testSessionCode = "123"
)

// testHash is md5("abc" + "123") = md5("abc123").
const testHash = "e99a18c428cb38d5f260853678922e03"

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := New(Config{
		PublicKey:   testPublicKey,
		PrivateKey:  testPrivateKey,
		SessionCode: testSessionCode,
	}, WithBaseURL(server.URL+"/"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client, server
}

func TestHashPrivateKey(t *testing.T) {
	if got := hashPrivateKey(testPrivateKey, testSessionCode); got != testHash {
		t.Errorf("hashPrivateKey = %q, want %q", got, testHash)
	}
}

func TestAuthenticate(t *testing.T) {
	var requests []authRequest

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/autorizacion-acceso/generar" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/autorizacion-acceso/generar")
		}
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		requests = append(requests, req)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","codigo_autorizacion":"token-1"}`))
	}))

	code, err := client.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if code != "token-1" {
		t.Errorf("code = %q, want token-1", code)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	want := authRequest{
		PublicKey:   testPublicKey,
		PrivateKey:  testHash,
		SessionCode: testSessionCode,
	}
	if requests[0] != want {
		t.Errorf("request = %+v, want %+v", requests[0], want)
	}

	// doAuthenticated reuses the cached code without hitting the auth API
	// again; Authenticate itself always forces a refresh.
	code, fresh, err := client.authorizationCode(context.Background())
	if err != nil {
		t.Fatalf("authorizationCode: %v", err)
	}
	if code != "token-1" || fresh {
		t.Errorf("authorizationCode = (%q, %t), want (token-1, false)", code, fresh)
	}
	if len(requests) != 1 {
		t.Errorf("requests = %d, want 1 (cached code reused)", len(requests))
	}
}

func TestAuthenticateInvalidCredentials(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"1","mensaje":"credenciales invalidas","codigo_autorizacion":null}`))
	}))

	_, err := client.Authenticate(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("error = %v, want *APIError", err)
	}
}

func TestAuthenticateEmptyCode(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","codigo_autorizacion":null}`))
	}))

	_, err := client.Authenticate(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("error = %v, want ErrUnauthorized", err)
	}
}

func TestTokenExpiryTriggersRegeneration(t *testing.T) {
	var authCalls int

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","codigo_autorizacion":"token-` + string(rune('0'+authCalls)) + `"}`))
	}))

	ctx := context.Background()
	if _, _, err := client.authorizationCode(ctx); err != nil {
		t.Fatalf("authorizationCode: %v", err)
	}
	if authCalls != 1 {
		t.Fatalf("authCalls = %d, want 1", authCalls)
	}

	// Simulate a token nearing expiry: backdate it past the refresh margin.
	client.mu.Lock()
	client.tokenAcquiredAt = time.Now().Add(-(tokenLifetime - tokenRefreshMargin) - time.Second)
	client.mu.Unlock()

	code, _, err := client.authorizationCode(ctx)
	if err != nil {
		t.Fatalf("authorizationCode: %v", err)
	}
	if authCalls != 2 {
		t.Errorf("authCalls = %d, want 2 (expired token regenerated)", authCalls)
	}
	if code != "token-2" {
		t.Errorf("code = %q, want token-2", code)
	}
}

// countingAuthEndpoint serves the auth endpoint, incrementing counters, and
// delegates any other path to next with the current auth-call count.
func countingAuthEndpoint(t *testing.T, authCalls *int, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/autorizacion-acceso/generar" {
			*authCalls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","codigo_autorizacion":"token-` + string(rune('0'+*authCalls)) + `"}`))
			return
		}
		next(w, r)
	})
}

type testAuthedRequest struct {
	authFields
	Origin string `json:"origen,omitempty"`
}

type testAuthedResponse struct {
	baseResponse
	Datos []string `json:"datos"`
}

func TestDoAuthenticatedInjectsCredentials(t *testing.T) {
	var gotOriginReq testAuthedRequest

	client, _ := newTestClient(t, countingAuthEndpoint(t, new(int), func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotOriginReq); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","datos":["a"]}`))
	}))

	req := testAuthedRequest{Origin: "ASU"}
	var out testAuthedResponse
	if err := client.doAuthenticated(context.Background(), "/envios/ciudades", &req, &out); err != nil {
		t.Fatalf("doAuthenticated: %v", err)
	}

	if gotOriginReq.PublicKey != testPublicKey {
		t.Errorf("clave_publica = %q, want %q", gotOriginReq.PublicKey, testPublicKey)
	}
	if gotOriginReq.AuthorizationCode != "token-1" {
		t.Errorf("codigo_autorizacion = %q, want token-1", gotOriginReq.AuthorizationCode)
	}
	if gotOriginReq.Origin != "ASU" {
		t.Errorf("origen = %q, want ASU", gotOriginReq.Origin)
	}
	if len(out.Datos) != 1 || out.Datos[0] != "a" {
		t.Errorf("datos = %v, want [a]", out.Datos)
	}
}

func TestDoAuthenticatedRetriesWithStaleToken(t *testing.T) {
	authCalls := 0
	endpointCalls := 0
	var endpointTokens []string

	client, _ := newTestClient(t, countingAuthEndpoint(t, &authCalls, func(w http.ResponseWriter, r *http.Request) {
		endpointCalls++
		var req testAuthedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		endpointTokens = append(endpointTokens, req.AuthorizationCode)

		w.Header().Set("Content-Type", "application/json")
		if endpointCalls == 1 {
			// Simulate the cached token having expired server-side.
			_, _ = w.Write([]byte(`{"codigo":"8","mensaje":"codigo de autorizacion invalido"}`))
			return
		}
		_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","datos":["a"]}`))
	}))

	// Prime the cache so the first endpoint call reuses a stale code.
	if _, err := client.Authenticate(context.Background()); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	req := testAuthedRequest{Origin: "ASU"}
	var out testAuthedResponse
	if err := client.doAuthenticated(context.Background(), "/envios/ciudades", &req, &out); err != nil {
		t.Fatalf("doAuthenticated: %v", err)
	}

	if endpointCalls != 2 {
		t.Errorf("endpointCalls = %d, want 2 (retry after re-auth)", endpointCalls)
	}
	if authCalls != 2 {
		t.Errorf("authCalls = %d, want 2", authCalls)
	}
	if len(endpointTokens) != 2 || endpointTokens[0] != "token-1" || endpointTokens[1] != "token-2" {
		t.Errorf("endpointTokens = %v, want [token-1 token-2]", endpointTokens)
	}
	if len(out.Datos) != 1 || out.Datos[0] != "a" {
		t.Errorf("datos = %v, want [a]", out.Datos)
	}
}

func TestDoAuthenticatedNoRetryWithFreshToken(t *testing.T) {
	authCalls := 0
	endpointCalls := 0

	client, _ := newTestClient(t, countingAuthEndpoint(t, &authCalls, func(w http.ResponseWriter, r *http.Request) {
		endpointCalls++
		w.Header().Set("Content-Type", "application/json")
		// Simulate a business error (not a token problem) with a fresh token.
		_, _ = w.Write([]byte(`{"codigo":"5","mensaje":"parametros invalidos"}`))
	}))

	req := testAuthedRequest{Origin: "BAD"}
	var out testAuthedResponse
	err := client.doAuthenticated(context.Background(), "/envios/ciudades", &req, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "5" {
		t.Errorf("error = %v, want APIError code 5", err)
	}
	if endpointCalls != 1 {
		t.Errorf("endpointCalls = %d, want 1 (no retry with fresh token)", endpointCalls)
	}
	if authCalls != 1 {
		t.Errorf("authCalls = %d, want 1", authCalls)
	}
}

func TestDoAuthenticatedSingleAuthForConcurrentCalls(t *testing.T) {
	authCalls := 0

	client, _ := newTestClient(t, countingAuthEndpoint(t, &authCalls, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"codigo":"0","mensaje":"OK","datos":[]}`))
	}))

	const goroutines = 8
	var wg sync.WaitGroup
	errs := make([]error, goroutines)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := testAuthedRequest{Origin: "ASU"}
			var out testAuthedResponse
			errs[i] = client.doAuthenticated(context.Background(), "/envios/ciudades", &req, &out)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}
	if authCalls != 1 {
		t.Errorf("authCalls = %d, want 1 (single-flight authentication)", authCalls)
	}
}
