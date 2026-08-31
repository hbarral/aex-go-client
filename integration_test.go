//go:build integration

package aex

import (
	"context"
	"os"
	"testing"
	"time"
)

// integrationConfig builds the sandbox credentials from the environment.
// Set AEX_PUBLIC_KEY, AEX_PRIVATE_KEY, and AEX_SESSION_CODE, then run:
//
//	go test -tags integration -run Integration -v
func integrationConfig(t *testing.T) Config {
	t.Helper()

	config := Config{
		PublicKey:   os.Getenv("AEX_PUBLIC_KEY"),
		PrivateKey:  os.Getenv("AEX_PRIVATE_KEY"),
		SessionCode: os.Getenv("AEX_SESSION_CODE"),
		Sandbox:     true,
	}
	if config.PublicKey == "" || config.PrivateKey == "" || config.SessionCode == "" {
		t.Skip("AEX_PUBLIC_KEY, AEX_PRIVATE_KEY, or AEX_SESSION_CODE not set")
	}
	return config
}

func TestIntegrationAuthenticate(t *testing.T) {
	client, err := New(integrationConfig(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	code, err := client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if code == "" {
		t.Fatal("Authenticate returned an empty authorization code")
	}
}

func TestIntegrationCities(t *testing.T) {
	client, err := New(integrationConfig(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cities, err := client.Cities(ctx, CitiesParams{})
	if err != nil {
		t.Fatalf("Cities: %v", err)
	}
	if len(cities) == 0 {
		t.Fatal("Cities returned no coverage cities")
	}
	t.Logf("coverage cities: %d (first: %s)", len(cities), cities[0].Name)
}
