package aex

import (
	"context"
	"errors"
	"testing"
)

func TestCities(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "ciudades.json", &captured))

	cities, err := client.Cities(context.Background(), CitiesParams{Origin: "ASU"})
	if err != nil {
		t.Fatalf("Cities: %v", err)
	}

	if len(cities) != 2 {
		t.Fatalf("len(cities) = %d, want 2", len(cities))
	}
	want := City{
		Code:           "ASU",
		Name:           "Asuncion",
		DepartmentCode: "11",
		DepartmentName: "Asuncion",
		CountryCode:    "PY",
		CountryName:    "Paraguay",
	}
	if cities[0] != want {
		t.Errorf("cities[0] = %+v, want %+v", cities[0], want)
	}
	if cities[1].Code != "CDE" || cities[1].DepartmentName != "Alto Parana" {
		t.Errorf("cities[1] = %+v", cities[1])
	}

	if captured["clave_publica"] != testPublicKey {
		t.Errorf("clave_publica = %v, want %q", captured["clave_publica"], testPublicKey)
	}
	if _, ok := captured["codigo_autorizacion"]; !ok {
		t.Error("codigo_autorizacion missing from request")
	}
	if captured["origen"] != "ASU" {
		t.Errorf("origen = %v, want ASU", captured["origen"])
	}
}

func TestCitiesAPIError(t *testing.T) {
	client, _ := newTestClient(t, errorEndpoint(t, "3", "sin cobertura"))

	_, err := client.Cities(context.Background(), CitiesParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *APIError", err)
	}
	if apiErr.Code != "3" {
		t.Errorf("Code = %q, want 3", apiErr.Code)
	}
	if apiErr.Endpoint != citiesPath {
		t.Errorf("Endpoint = %q, want %q", apiErr.Endpoint, citiesPath)
	}
}
