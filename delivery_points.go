package aex

import (
	"context"
	"fmt"
)

const deliveryPointsPath = "envios/puntos_entrega"

// DeliveryPointsParams are the parameters of the Shipments Delivery Points
// method.
type DeliveryPointsParams struct {
	authFields
	// ServiceTypeID identifies the service type to be contracted; the
	// value is provided by AEX.
	ServiceTypeID FlexInt `json:"id_tipo_servicio"`
	// Origin is the origin city code.
	Origin string `json:"origen"`
	// Destino is the destination city code.
	Destino string `json:"destino"`
}

func (p DeliveryPointsParams) validate() error {
	if p.ServiceTypeID == 0 {
		return fmt.Errorf("aex: delivery points: ServiceTypeID is required")
	}
	if p.Origin == "" {
		return fmt.Errorf("aex: delivery points: Origin is required")
	}
	if p.Destino == "" {
		return fmt.Errorf("aex: delivery points: Destino is required")
	}
	return nil
}

// DeliveryPoint is a CAC (service center) or ELOCKER (self-service
// terminal) point where packages can be deposited or picked up.
type DeliveryPoint struct {
	// ID is the delivery point identifier. The API encodes it
	// inconsistently (string or number) depending on the endpoint, so it
	// is a FlexInt.
	ID FlexInt `json:"id"`
	// Type is PointTypeCAC or PointTypeELocker.
	Type string `json:"tipo"`
	// Pickup reports whether package deposit is enabled.
	Pickup FlexBool `json:"pickup"`
	// Delivery reports whether package delivery is enabled.
	Delivery FlexBool `json:"entrega"`
	// Latitude is the point latitude.
	Latitude FlexFloat `json:"latitud"`
	// Longitude is the point longitude.
	Longitude FlexFloat `json:"longitud"`
	// Phone is the delivery point phone number.
	Phone string `json:"telefono"`
	// Address is the delivery point address.
	Address string `json:"direccion"`
	// Outsourced reports whether the point is outsourced rather than an
	// AEX-owned office.
	Outsourced FlexBool `json:"tercerizado"`
	// CityCode is the code of the city where the point is located.
	CityCode string `json:"codigo_ciudad"`
	// Name is the delivery point name.
	Name string `json:"punto_entrega"`
	// Default reports whether the point is the default for the service
	// type.
	Default FlexBool `json:"predeterminado"`
	// OpeningHours describes when the point is available.
	OpeningHours string `json:"horario_atencion"`
}

// DeliveryPoints returns the delivery points available for the given
// service type and route.
func (c *Client) DeliveryPoints(ctx context.Context, params DeliveryPointsParams) ([]DeliveryPoint, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	var resp struct {
		baseResponse
		Datos []DeliveryPoint `json:"datos"`
	}
	if err := c.doAuthenticated(ctx, deliveryPointsPath, &params, &resp); err != nil {
		return nil, err
	}
	return resp.Datos, nil
}
