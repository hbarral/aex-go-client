package aex

import (
	"context"
	"testing"
	"time"
)

func TestTracking(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "tracking.json", &captured))

	events, err := client.Tracking(context.Background(), TrackingParams{GuideNumber: "A002866303"})
	if err != nil {
		t.Fatalf("Tracking: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}

	first := events[0]
	if first.GuideNumber != "A002866303" {
		t.Errorf("GuideNumber = %q", first.GuideNumber)
	}
	wantDate := time.Date(2025, time.June, 6, 12, 34, 56, 0, time.UTC)
	if !first.Date.Time.Equal(wantDate) {
		t.Errorf("Date = %v, want %v", first.Date.Time, wantDate)
	}
	if first.StatusCode != "E" || first.Status != "Entregado" {
		t.Errorf("first = %+v", first)
	}
	if first.EventTypeCode != "W5" || first.EventType != "Entrega Realizada" {
		t.Errorf("first = %+v", first)
	}
	if first.Observation != "Paquete entregado correctamente" {
		t.Errorf("Observation = %q", first.Observation)
	}

	// A null observation must decode as the empty string, not fail.
	if events[1].Observation != "" {
		t.Errorf("Observation = %q, want empty for null", events[1].Observation)
	}

	if captured["numero_guia"] != "A002866303" {
		t.Errorf("numero_guia = %v", captured["numero_guia"])
	}
	if _, ok := captured["codigo_operacion"]; ok {
		t.Error("codigo_operacion should be omitted when empty")
	}
}

func TestTrackingByOperationCode(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "tracking.json", &captured))

	_, err := client.Tracking(context.Background(), TrackingParams{OperationCode: "PEDIDO-12345"})
	if err != nil {
		t.Fatalf("Tracking: %v", err)
	}
	if captured["codigo_operacion"] != "PEDIDO-12345" {
		t.Errorf("codigo_operacion = %v", captured["codigo_operacion"])
	}
}

func TestTrackingBothSelectorsSent(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "tracking.json", &captured))

	// The API gives priority to the guide number when both are sent; the
	// client sends both as documented.
	_, err := client.Tracking(context.Background(), TrackingParams{
		GuideNumber:   "A002866303",
		OperationCode: "PEDIDO-12345",
	})
	if err != nil {
		t.Fatalf("Tracking: %v", err)
	}
	if captured["numero_guia"] != "A002866303" {
		t.Errorf("numero_guia = %v", captured["numero_guia"])
	}
	if captured["codigo_operacion"] != "PEDIDO-12345" {
		t.Errorf("codigo_operacion = %v", captured["codigo_operacion"])
	}
}

func TestTrackingValidation(t *testing.T) {
	client, _ := newTestClient(t, fixtureEndpoint(t, "tracking.json", nil))

	_, err := client.Tracking(context.Background(), TrackingParams{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
