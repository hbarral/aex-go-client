package aex

import (
	"context"
	"fmt"
)

const calculatePath = "envios/calcular"

// Package describes a package to be shipped. It is used both for quotes
// (Calculate) and service requests (RequestService); fields that do not
// apply to a call are omitted when zero.
type Package struct {
	// Description holds the product details.
	Description string `json:"descripcion,omitempty"`
	// ExternalCode is an article or package identifier code.
	ExternalCode string `json:"codigo_externo,omitempty"`
	// Quantity is the number of packages with the same characteristics.
	// The API assumes 1 when omitted.
	Quantity int `json:"cantidad,omitempty"`
	// Weight is the package weight in kilograms.
	Weight float64 `json:"peso"`
	// Length is the package length in centimeters.
	Length float64 `json:"largo"`
	// Height is the package height in centimeters.
	Height float64 `json:"alto"`
	// Width is the package width in centimeters.
	Width float64 `json:"ancho"`
	// Value is the declared value of the merchandise in PYG.
	Value float64 `json:"valor,omitempty"`
	// ProductTypeCode identifies the product type for special type
	// applications, agreed with AEX beforehand.
	ProductTypeCode string `json:"codigo_tipo_producto,omitempty"`
	// ProductCode (SKU) relates the package to inventory. Only applies to
	// services where AEX stores the products to be shipped.
	ProductCode string `json:"codigo_producto,omitempty"`
}

// CalculateParams are the parameters of the Shipments Calculate method.
type CalculateParams struct {
	authFields
	// Origin is the origin city code.
	Origin string `json:"origen"`
	// Destino is the destination city code.
	Destino string `json:"destino"`
	// Packages is the list of packages to be shipped.
	Packages []Package `json:"paquetes"`
	// LoadType is the cargo type code: LoadPackage ("P", the default) or
	// LoadDocument ("D"). Omitted when empty.
	LoadType string `json:"codigo_tipo_carga,omitempty"`
}

func (p CalculateParams) validate() error {
	if p.Origin == "" {
		return fmt.Errorf("aex: calculate: Origin is required")
	}
	if p.Destino == "" {
		return fmt.Errorf("aex: calculate: Destino is required")
	}
	if len(p.Packages) == 0 {
		return fmt.Errorf("aex: calculate: at least one package is required")
	}
	return nil
}

// ServiceQuote is a service available for a shipment, with its freight
// cost and additional services, as returned by Calculate.
type ServiceQuote struct {
	// ServiceTypeID is the AEX service identifier.
	ServiceTypeID int `json:"id_tipo_servicio"`
	// ServiceType is the service name.
	ServiceType string `json:"tipo_servicio"`
	// Description is the detailed service description.
	Description string `json:"descripcion"`
	// DeliveryTime is the estimated maximum delivery time in hours,
	// counted from cargo pickup.
	DeliveryTime int `json:"tiempo_entrega"`
	// IncludesPickup reports whether the service requires pickup; when
	// false the client leaves packages at a delivery point.
	IncludesPickup TFBool `json:"incluye_pickup"`
	// IncludesDelivery reports whether the service requires delivery;
	// when false the client picks packages up at a delivery point.
	IncludesDelivery TFBool `json:"incluye_envio"`
	// FreightCost is the freight cost in guaraníes.
	FreightCost float64 `json:"costo_flete"`
	// AdditionalServices lists extra services besides freight (e.g.,
	// insurance).
	AdditionalServices []AdditionalService `json:"adicionales"`
}

// Calculate returns the shipping cost for each available service according
// to the packages, origin, and destination.
func (c *Client) Calculate(ctx context.Context, params CalculateParams) ([]ServiceQuote, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	var resp struct {
		baseResponse
		Datos []ServiceQuote `json:"datos"`
	}
	if err := c.doAuthenticated(ctx, calculatePath, &params, &resp); err != nil {
		return nil, err
	}
	return resp.Datos, nil
}
