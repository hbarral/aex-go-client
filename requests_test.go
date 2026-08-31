package aex

import (
	"context"
	"errors"
	"testing"
)

func TestRequestService(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "solicitar_servicio.json", &captured))

	offer, err := client.RequestService(context.Background(), RequestServiceParams{
		Origin:           "ASU",
		Destino:          "CDE",
		OperationCode:    "PEDIDO-12345",
		CollectionAmount: 150000,
		Packages: []Package{{
			Description: "Camiseta",
			Weight:      1.5,
			Length:      30,
			Height:      10,
			Width:       20,
			Value:       100000,
		}},
	})
	if err != nil {
		t.Fatalf("RequestService: %v", err)
	}

	if offer.ID != 123 {
		t.Errorf("offer.ID = %d, want 123", offer.ID)
	}
	if len(offer.Conditions) != 2 {
		t.Fatalf("len(Conditions) = %d, want 2", len(offer.Conditions))
	}

	cond := offer.Conditions[0]
	if cond.ServiceTypeID != 3 || cond.ServiceType != "Moto" {
		t.Errorf("cond = %+v", cond)
	}
	if !cond.IncludesPickup || !cond.IncludesDelivery {
		t.Errorf("t/f booleans: pickup %t delivery %t", cond.IncludesPickup, cond.IncludesDelivery)
	}
	if len(cond.AdditionalServices) != 2 {
		t.Fatalf("len(AdditionalServices) = %d, want 2", len(cond.AdditionalServices))
	}

	insurance := cond.AdditionalServices[0]
	if insurance.ID != 1 || insurance.Name != "Seguro" || insurance.Cost != 500 {
		t.Errorf("insurance = %+v", insurance)
	}
	if insurance.Mandatory || insurance.Collection || !insurance.Preselected {
		t.Errorf("insurance flags: obligatorio %t cobranza %t preseleccionado %t",
			insurance.Mandatory, insurance.Collection, insurance.Preselected)
	}
	if insurance.Value != 100000 {
		t.Errorf("insurance.Value = %v, want 100000", insurance.Value)
	}
	if insurance.Return != nil {
		t.Errorf("insurance.Return = %+v, want nil", insurance.Return)
	}

	cod := cond.AdditionalServices[1]
	if !cod.Collection {
		t.Error("cobranza additional should report Collection true")
	}
	if cod.Return == nil {
		t.Fatal("devolucion should decode for collection additional")
	}
	if cod.Return.ServiceTypeID != 5 || cod.Return.ServiceType != "Moto retorno" {
		t.Errorf("Return = %+v", cod.Return)
	}
	if cod.Return.Cost != 20000 || cod.Return.LoadType != LoadDocument {
		t.Errorf("Return = %+v", cod.Return)
	}

	if len(cond.DeliveryPoints) != 1 || cond.DeliveryPoints[0].ID != 17 {
		t.Errorf("DeliveryPoints = %+v", cond.DeliveryPoints)
	}

	// Request payload checks.
	if captured["codigo_operacion"] != "PEDIDO-12345" {
		t.Errorf("codigo_operacion = %v", captured["codigo_operacion"])
	}
	if captured["importe_cobro"] != float64(150000) {
		t.Errorf("importe_cobro = %v, want 150000", captured["importe_cobro"])
	}
	if _, ok := captured["codigo_tipo_carga"]; ok {
		t.Error("codigo_tipo_carga should be omitted when LoadType is empty")
	}
}

func TestRequestServiceArrayDatos(t *testing.T) {
	// The documentation is ambiguous about the shape of datos; the client
	// tolerates a one-element array too.
	client, _ := newTestClient(t, fixtureEndpoint(t, "solicitar_servicio_array.json", nil))

	offer, err := client.RequestService(context.Background(), RequestServiceParams{
		Origin:   "ASU",
		Destino:  "CDE",
		Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}},
	})
	if err != nil {
		t.Fatalf("RequestService: %v", err)
	}
	if offer.ID != 456 {
		t.Errorf("offer.ID = %d, want 456", offer.ID)
	}
	if len(offer.Conditions) != 1 || offer.Conditions[0].ServiceTypeID != 3 {
		t.Errorf("offer = %+v", offer)
	}
}

func TestRequestServiceNullDatos(t *testing.T) {
	client, _ := newTestClient(t, errorEndpoint(t, "0", "OK"))

	_, err := client.RequestService(context.Background(), RequestServiceParams{
		Origin:   "ASU",
		Destino:  "CDE",
		Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}},
	})
	if err == nil {
		t.Fatal("expected error for null datos, got nil")
	}
}

func TestRequestServiceValidation(t *testing.T) {
	tests := []struct {
		name   string
		params RequestServiceParams
	}{
		{"missing origin", RequestServiceParams{Destino: "CDE", Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}}}},
		{"missing destino", RequestServiceParams{Origin: "ASU", Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}}}},
		{"no packages", RequestServiceParams{Origin: "ASU", Destino: "CDE"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, fixtureEndpoint(t, "solicitar_servicio.json", nil))
			_, err := client.RequestService(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestRequestServiceAPIError(t *testing.T) {
	client, _ := newTestClient(t, errorEndpoint(t, "9", "oferta expirada"))

	_, err := client.RequestService(context.Background(), RequestServiceParams{
		Origin:   "ASU",
		Destino:  "CDE",
		Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}},
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Endpoint != requestServicePath {
		t.Errorf("Endpoint = %q, want %q", apiErr.Endpoint, requestServicePath)
	}
}
