package aex

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestCancelService(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "cancelar.json", &captured))

	if err := client.CancelService(context.Background(), "A002866303"); err != nil {
		t.Fatalf("CancelService: %v", err)
	}
	if captured["numero_guia"] != "A002866303" {
		t.Errorf("numero_guia = %v", captured["numero_guia"])
	}
}

func TestCancelServiceValidation(t *testing.T) {
	client, _ := newTestClient(t, fixtureEndpoint(t, "cancelar.json", nil))

	if err := client.CancelService(context.Background(), ""); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestCancelServiceNotCancellable(t *testing.T) {
	client, _ := newTestClient(t, errorEndpoint(t, "5", "la guia ya tuvo gestion operativa"))

	err := client.CancelService(context.Background(), "A002866303")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Code != "5" {
		t.Errorf("Code = %q, want 5", apiErr.Code)
	}
}

func TestModifyGuide(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, countingAuthEndpoint(t, new(int), func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/actualizar-guia.php" {
			t.Errorf("path = %q, want /actualizar-guia.php", r.URL.Path)
		}
		if accion := r.URL.Query().Get("accion"); accion != "modificar" {
			t.Errorf("accion = %q, want modificar", accion)
		}
		if err := jsonDecodeStrict(r, &captured); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, "actualizar_guia.json"))
	}))

	err := client.ModifyGuide(context.Background(), ModifyGuideParams{
		GuideNumber: "A002866303",
		Modify: GuideModifications{
			PickupInstructions:   "Llamar antes de llegar",
			DeliveryInstructions: "Entregar en recepcion",
		},
	})
	if err != nil {
		t.Fatalf("ModifyGuide: %v", err)
	}

	if captured["numero_guia"] != "A002866303" {
		t.Errorf("numero_guia = %v", captured["numero_guia"])
	}
	modify, _ := captured["modificar"].(map[string]any)
	if modify == nil {
		t.Fatal("modificar missing from request")
	}
	if modify["indicacion_pickup"] != "Llamar antes de llegar" {
		t.Errorf("indicacion_pickup = %v", modify["indicacion_pickup"])
	}
	if modify["indicacion_entrega"] != "Entregar en recepcion" {
		t.Errorf("indicacion_entrega = %v", modify["indicacion_entrega"])
	}
}

func TestModifyGuideOmitsEmptyInstruction(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "actualizar_guia.json", &captured))

	err := client.ModifyGuide(context.Background(), ModifyGuideParams{
		GuideNumber: "A002866303",
		Modify:      GuideModifications{PickupInstructions: "Solo pickup"},
	})
	if err != nil {
		t.Fatalf("ModifyGuide: %v", err)
	}
	modify, _ := captured["modificar"].(map[string]any)
	if _, ok := modify["indicacion_entrega"]; ok {
		t.Error("indicacion_entrega should be omitted when empty (API rejects empty values)")
	}
}

func TestModifyGuideValidation(t *testing.T) {
	client, _ := newTestClient(t, fixtureEndpoint(t, "actualizar_guia.json", nil))

	tests := []struct {
		name   string
		params ModifyGuideParams
	}{
		{"missing guide number", ModifyGuideParams{Modify: GuideModifications{PickupInstructions: "x"}}},
		{"no instructions", ModifyGuideParams{GuideNumber: "A002866303"}},
		{
			"pickup instructions too long",
			ModifyGuideParams{
				GuideNumber: "A002866303",
				Modify:      GuideModifications{PickupInstructions: strings.Repeat("x", instructionMaxLength+1)},
			},
		},
		{
			"delivery instructions too long",
			ModifyGuideParams{
				GuideNumber: "A002866303",
				Modify:      GuideModifications{DeliveryInstructions: strings.Repeat("x", instructionMaxLength+1)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.ModifyGuide(context.Background(), tt.params); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestIncidentManagement(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "gestion_novedad.json", &captured))

	managementDate := Date{}
	managementDate.Time = dateFrom(t, "2025-06-06")
	err := client.IncidentManagement(context.Background(), IncidentManagementParams{
		GuideNumber:    "A002866303",
		EventTypeCode:  "N1",
		ManagementDate: &managementDate,
		Observations:   "Se reprogramo la entrega",
	})
	if err != nil {
		t.Fatalf("IncidentManagement: %v", err)
	}

	if captured["numero_guia"] != "A002866303" {
		t.Errorf("numero_guia = %v", captured["numero_guia"])
	}
	if captured["codigo_tipo_evento"] != "N1" {
		t.Errorf("codigo_tipo_evento = %v", captured["codigo_tipo_evento"])
	}
	if captured["fecha_gestion"] != "2025-06-06" {
		t.Errorf("fecha_gestion = %v, want 2025-06-06", captured["fecha_gestion"])
	}
	if captured["observaciones"] != "Se reprogramo la entrega" {
		t.Errorf("observaciones = %v", captured["observaciones"])
	}
}

func TestIncidentManagementValidation(t *testing.T) {
	client, _ := newTestClient(t, fixtureEndpoint(t, "gestion_novedad.json", nil))

	date := Date{}
	date.Time = dateFrom(t, "2025-06-06")
	tests := []struct {
		name   string
		params IncidentManagementParams
	}{
		{"missing guide number", IncidentManagementParams{EventTypeCode: "N1", ManagementDate: &date}},
		{"missing event type", IncidentManagementParams{GuideNumber: "A002866303", ManagementDate: &date}},
		{"missing date", IncidentManagementParams{GuideNumber: "A002866303", EventTypeCode: "N1"}},
		{
			"zero date",
			IncidentManagementParams{GuideNumber: "A002866303", EventTypeCode: "N1", ManagementDate: &Date{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.IncidentManagement(context.Background(), tt.params); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}
