package aex

import (
	"context"
	"testing"
	"time"
)

func validConfirmParams() ConfirmServiceParams {
	birth := Date{Time: time.Date(1990, time.March, 15, 0, 0, 0, 0, time.UTC)}
	return ConfirmServiceParams{
		RequestID:     123,
		ServiceTypeID: 3,
		Pickup: &Location{
			AddressCode:  "ADDR-1",
			MainStreet:   "Av. Brasil",
			CrossStreet1: "Calle Debussy",
			CityCode:     "ASU",
		},
		Recipient: &Party{
			DocumentNumber: "1234567",
			Name:           "Juan",
			LastName:       "Perez",
			Email:          "juan@example.com",
			BirthDate:      &birth,
			Phones:         []Phone{{Number: 98111222, Name: "celular"}},
		},
		Delivery:      NewDeliveryPointLocation(17),
		AdditionalIDs: []int{1},
		PaymentMethod: PaymentDestination,
	}
}

func TestConfirmService(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "confirmar_servicio.json", &captured))

	if err := client.ConfirmService(context.Background(), validConfirmParams()); err != nil {
		t.Fatalf("ConfirmService: %v", err)
	}

	if captured["id_solicitud"] != float64(123) {
		t.Errorf("id_solicitud = %v, want 123", captured["id_solicitud"])
	}
	if captured["id_tipo_servicio"] != float64(3) {
		t.Errorf("id_tipo_servicio = %v, want 3", captured["id_tipo_servicio"])
	}

	pickup, _ := captured["pickup"].(map[string]any)
	if pickup == nil {
		t.Fatal("pickup missing from request")
	}
	if pickup["codigo"] != "ADDR-1" || pickup["calle_principal"] != "Av. Brasil" {
		t.Errorf("pickup = %v", pickup)
	}
	if pickup["calle_transversal_1"] != "Calle Debussy" || pickup["codigo_ciudad"] != "ASU" {
		t.Errorf("pickup = %v", pickup)
	}

	delivery, _ := captured["entrega"].(map[string]any)
	if delivery == nil {
		t.Fatal("entrega missing from request")
	}
	if delivery["id_punto_entrega"] != float64(17) {
		t.Errorf("entrega = %v, want id_punto_entrega 17", delivery)
	}
	if _, ok := delivery["calle_principal"]; ok {
		t.Error("address fields should be omitted when DeliveryPointID is set")
	}

	recipient, _ := captured["destinatario"].(map[string]any)
	if recipient == nil {
		t.Fatal("destinatario missing from request")
	}
	if recipient["numero_documento"] != "1234567" || recipient["nombre"] != "Juan" {
		t.Errorf("destinatario = %v", recipient)
	}
	if recipient["fecha_nacimiento"] != "1990-03-15" {
		t.Errorf("fecha_nacimiento = %v, want 1990-03-15", recipient["fecha_nacimiento"])
	}
	phones, _ := recipient["telefonos"].([]any)
	if len(phones) != 1 {
		t.Fatalf("telefonos = %v", recipient["telefonos"])
	}
	phone, _ := phones[0].(map[string]any)
	if phone["numero"] != float64(98111222) {
		t.Errorf("telefono = %v", phone)
	}

	if _, ok := captured["remitente"]; ok {
		t.Error("remitente should be omitted when Sender is nil")
	}

	ids, _ := captured["adicionales"].([]any)
	if len(ids) != 1 || ids[0] != float64(1) {
		t.Errorf("adicionales = %v, want [1]", captured["adicionales"])
	}
	if captured["codigo_forma_pago"] != PaymentDestination {
		t.Errorf("codigo_forma_pago = %v", captured["codigo_forma_pago"])
	}
}

func TestConfirmServiceCustomGuide(t *testing.T) {
	var captured map[string]any

	client, _ := newTestClient(t, fixtureEndpoint(t, "confirmar_servicio.json", &captured))

	params := validConfirmParams()
	params.SerialNumber = "ML"
	params.GuideNumber = "100000456"
	if err := client.ConfirmService(context.Background(), params); err != nil {
		t.Fatalf("ConfirmService: %v", err)
	}
	if captured["numero_serie"] != "ML" || captured["numero_guia"] != "100000456" {
		t.Errorf("numero_serie/numero_guia = %v/%v", captured["numero_serie"], captured["numero_guia"])
	}
}

func TestConfirmServiceValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ConfirmServiceParams)
	}{
		{"missing request id", func(p *ConfirmServiceParams) { p.RequestID = 0 }},
		{"missing service type id", func(p *ConfirmServiceParams) { p.ServiceTypeID = 0 }},
		{"missing pickup", func(p *ConfirmServiceParams) { p.Pickup = nil }},
		{"missing delivery", func(p *ConfirmServiceParams) { p.Delivery = nil }},
		{"missing recipient", func(p *ConfirmServiceParams) { p.Recipient = nil }},
		{
			"pickup missing address code",
			func(p *ConfirmServiceParams) { p.Pickup.AddressCode = "" },
		},
		{
			"pickup missing main street",
			func(p *ConfirmServiceParams) { p.Pickup.MainStreet = "" },
		},
		{
			"delivery missing city code",
			func(p *ConfirmServiceParams) {
				p.Delivery = &Location{AddressCode: "ADDR-2", MainStreet: "Calle", CrossStreet1: "Esquina"}
			},
		},
		{
			"recipient without phones",
			func(p *ConfirmServiceParams) { p.Recipient.Phones = nil },
		},
		{
			"recipient missing email",
			func(p *ConfirmServiceParams) { p.Recipient.Email = "" },
		},
		{
			"sender missing document",
			func(p *ConfirmServiceParams) {
				p.Sender = &Party{Name: "Ana", Email: "ana@example.com", Phones: []Phone{{Number: 1}}}
			},
		},
		{
			"serial without guide",
			func(p *ConfirmServiceParams) { p.SerialNumber = "ML" },
		},
		{
			"guide without serial",
			func(p *ConfirmServiceParams) { p.GuideNumber = "100000456" },
		},
		{
			"short guide number",
			func(p *ConfirmServiceParams) {
				p.SerialNumber = "ML"
				p.GuideNumber = "12345"
			},
		},
		{
			"guide with invalid characters",
			func(p *ConfirmServiceParams) {
				p.SerialNumber = "ML"
				p.GuideNumber = "100-456"
			},
		},
		{
			"serial too long",
			func(p *ConfirmServiceParams) {
				p.SerialNumber = "ABCD"
				p.GuideNumber = "100000456"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, fixtureEndpoint(t, "confirmar_servicio.json", nil))
			params := validConfirmParams()
			tt.mutate(&params)
			if err := client.ConfirmService(context.Background(), params); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}
