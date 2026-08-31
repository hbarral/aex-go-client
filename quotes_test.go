package aex

import (
	"context"
	"errors"
	"testing"
)

func TestCalculate(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "calcular.json", &captured))

	quotes, err := client.Calculate(context.Background(), CalculateParams{
		Origin:  "ASU",
		Destino: "CDE",
		Packages: []Package{{
			Description: "Camiseta",
			Quantity:    2,
			Weight:      1.5,
			Length:      30,
			Height:      10,
			Width:       20,
			Value:       100000,
		}},
	})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("len(quotes) = %d, want 2", len(quotes))
	}

	first := quotes[0]
	if first.ServiceTypeID != 3 || first.ServiceType != "Moto" {
		t.Errorf("first = %+v", first)
	}
	if first.DeliveryTime != 6 {
		t.Errorf("DeliveryTime = %d, want 6", first.DeliveryTime)
	}
	if !first.IncludesPickup || !first.IncludesDelivery {
		t.Errorf("t/f booleans: pickup %t delivery %t, want true true",
			first.IncludesPickup, first.IncludesDelivery)
	}
	if first.FreightCost != 25000.5 {
		t.Errorf("FreightCost = %v, want 25000.5", first.FreightCost)
	}
	if len(first.AdditionalServices) != 1 {
		t.Fatalf("len(AdditionalServices) = %d, want 1", len(first.AdditionalServices))
	}
	if got := first.AdditionalServices[0]; got.ID != 1 || got.Name != "Seguro" || got.Cost != 500 {
		t.Errorf("AdditionalServices[0] = %+v", got)
	}

	second := quotes[1]
	if second.IncludesPickup || second.IncludesDelivery {
		t.Errorf("t/f booleans: pickup %t delivery %t, want false false",
			second.IncludesPickup, second.IncludesDelivery)
	}

	// Request payload checks.
	if captured["origen"] != "ASU" || captured["destino"] != "CDE" {
		t.Errorf("origen/destino = %v/%v", captured["origen"], captured["destino"])
	}
	if _, ok := captured["codigo_tipo_carga"]; ok {
		t.Error("codigo_tipo_carga should be omitted when LoadType is empty")
	}
	packages, ok := captured["paquetes"].([]any)
	if !ok || len(packages) != 1 {
		t.Fatalf("paquetes = %v, want 1 package", captured["paquetes"])
	}
	pkg, _ := packages[0].(map[string]any)
	if pkg["peso"] != 1.5 || pkg["largo"] != float64(30) {
		t.Errorf("pkg = %v", pkg)
	}
	if pkg["cantidad"] != float64(2) {
		t.Errorf("cantidad = %v, want 2", pkg["cantidad"])
	}
}

func TestCalculateLoadTypeSent(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "calcular.json", &captured))

	_, err := client.Calculate(context.Background(), CalculateParams{
		Origin:   "ASU",
		Destino:  "CDE",
		Packages: []Package{{Weight: 0.2, Length: 24, Height: 1, Width: 32}},
		LoadType: LoadDocument,
	})
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if captured["codigo_tipo_carga"] != LoadDocument {
		t.Errorf("codigo_tipo_carga = %v, want %q", captured["codigo_tipo_carga"], LoadDocument)
	}
}

func TestCalculateValidation(t *testing.T) {
	tests := []struct {
		name   string
		params CalculateParams
	}{
		{"missing origin", CalculateParams{Destino: "CDE", Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}}}},
		{"missing destino", CalculateParams{Origin: "ASU", Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}}}},
		{"no packages", CalculateParams{Origin: "ASU", Destino: "CDE"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, fixtureEndpoint(t, "calcular.json", nil))
			_, err := client.Calculate(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestCalculateAPIError(t *testing.T) {
	client, _ := newTestClient(t, errorEndpoint(t, "7", "ciudad sin cobertura"))

	_, err := client.Calculate(context.Background(), CalculateParams{
		Origin:   "ASU",
		Destino:  "XXX",
		Packages: []Package{{Weight: 1, Length: 1, Height: 1, Width: 1}},
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Code != "7" || apiErr.Endpoint != calculatePath {
		t.Errorf("apiErr = %+v", apiErr)
	}
}
