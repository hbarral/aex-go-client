package aex

import (
	"context"
	"testing"
)

func TestInventory(t *testing.T) {
	client, _ := newTestClient(t, fixtureEndpoint(t, "inventario.json", nil))

	stock, err := client.Inventory(context.Background(), InventoryParams{})
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}

	if len(stock) != 2 {
		t.Fatalf("len(stock) = %d, want 2", len(stock))
	}
	want := ProductStock{ProductCode: "SKU-1", Stock: 12, Name: "Camiseta"}
	if stock[0] != want {
		t.Errorf("stock[0] = %+v, want %+v", stock[0], want)
	}
	if stock[1].Stock != 0 || stock[1].Name != "Taza" {
		t.Errorf("stock[1] = %+v", stock[1])
	}
}

func TestInventoryProductCodes(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "inventario.json", &captured))

	_, err := client.Inventory(context.Background(), InventoryParams{
		ProductCodes: []string{"SKU-1"},
	})
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}

	codes, ok := captured["codigos_producto"].([]any)
	if !ok || len(codes) != 1 || codes[0] != "SKU-1" {
		t.Errorf("codigos_producto = %v, want [SKU-1]", captured["codigos_producto"])
	}
}

func TestInventoryOmitsEmptyProductCodes(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "inventario.json", &captured))

	_, err := client.Inventory(context.Background(), InventoryParams{})
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if _, ok := captured["codigos_producto"]; ok {
		t.Error("codigos_producto should be omitted when empty")
	}
}
