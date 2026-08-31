package aex

import (
	"context"
	"testing"
)

func TestDeliveryPoints(t *testing.T) {
	client, _ := newTestClient(t, fixtureEndpoint(t, "puntos_entrega.json", nil))

	points, err := client.DeliveryPoints(context.Background(), DeliveryPointsParams{
		ServiceTypeID: 3,
		Origin:        "ASU",
		Destino:       "CDE",
	})
	if err != nil {
		t.Fatalf("DeliveryPoints: %v", err)
	}

	if len(points) != 2 {
		t.Fatalf("len(points) = %d, want 2", len(points))
	}

	// First point carries a string id, per this endpoint's documentation.
	first := points[0]
	if first.ID != 17 {
		t.Errorf("ID = %d, want 17 (string id decoded)", first.ID)
	}
	if first.Type != PointTypeCAC {
		t.Errorf("Type = %q, want %q", first.Type, PointTypeCAC)
	}
	if !first.Pickup || !first.Delivery || first.Outsourced || !first.Default {
		t.Errorf("flags = pickup %t entrega %t tercerizado %t predeterminado %t",
			first.Pickup, first.Delivery, first.Outsourced, first.Default)
	}
	if first.CityCode != "ASU" || first.Name != "CAC Centro" || first.Address != "Av. Example 123" {
		t.Errorf("first = %+v", first)
	}

	// Second point carries a numeric id; both encodings must decode.
	second := points[1]
	if second.ID != 42 {
		t.Errorf("ID = %d, want 42 (numeric id decoded)", second.ID)
	}
	if second.Type != PointTypeELocker {
		t.Errorf("Type = %q, want %q", second.Type, PointTypeELocker)
	}
	if !second.Pickup || second.Delivery || !second.Outsourced || second.Default {
		t.Errorf("flags = pickup %t entrega %t tercerizado %t predeterminado %t",
			second.Pickup, second.Delivery, second.Outsourced, second.Default)
	}
}

func TestDeliveryPointsValidation(t *testing.T) {
	tests := []struct {
		name   string
		params DeliveryPointsParams
	}{
		{"missing service type", DeliveryPointsParams{Origin: "ASU", Destino: "CDE"}},
		{"missing origin", DeliveryPointsParams{ServiceTypeID: 3, Destino: "CDE"}},
		{"missing destino", DeliveryPointsParams{ServiceTypeID: 3, Origin: "ASU"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, fixtureEndpoint(t, "puntos_entrega.json", nil))
			_, err := client.DeliveryPoints(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}
