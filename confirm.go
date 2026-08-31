package aex

import (
	"context"
	"fmt"
	"regexp"
)

const confirmServicePath = "envios/confirmar_servicio"

// Payment methods for codigo_forma_pago.
const (
	// PaymentCredit bills the service to the client's credit account (the
	// API default when the client is enabled for this modality).
	PaymentCredit = "C"
	// PaymentOrigin charges the service amount to the sender at origin.
	PaymentOrigin = "O"
	// PaymentDestination charges the service amount to the recipient at
	// destination, on top of any collection requested by the sender.
	PaymentDestination = "D"
)

// Document types for Party.DocumentType.
const (
	DocumentTypeRUC      = "RUC"
	DocumentTypeCIP      = "CIP"
	DocumentTypePassport = "PAS"
)

// Person status values for Party.PersonStatus.
const (
	// PersonNatural is a natural person (personeria "F").
	PersonNatural = "F"
	// PersonLegal is a legal entity (personeria "J").
	PersonLegal = "J"
)

// guideNumberRe validates a custom guide number: 6 to 18 alphanumeric
// characters (A-Z, 0-9).
var guideNumberRe = regexp.MustCompile(`^[A-Z0-9]{6,18}$`)

// serialNumberRe validates a custom serial number: 1 to 3 alphanumeric
// characters. Values other than the AEX-reserved ones must be registered
// with AEX in advance or the API will reject them.
var serialNumberRe = regexp.MustCompile(`^[A-Z0-9]{1,3}$`)

// Phone is a phone number of a sender or recipient.
type Phone struct {
	// Number includes area codes.
	Number int `json:"numero"`
	// Name labels the phone.
	Name string `json:"denominacion,omitempty"`
	// Comments describes communication conditions for the number; the API
	// truncates it to 250 characters.
	Comments string `json:"comentarios,omitempty"`
}

// Party is a sender (remitente) or recipient (destinatario).
type Party struct {
	// Code uniquely identifies the client in the external application.
	Code string `json:"codigo,omitempty"`
	// DocumentType is DocumentTypeRUC, DocumentTypeCIP, or
	// DocumentTypePassport.
	DocumentType string `json:"tipo_documento,omitempty"`
	// DocumentNumber is the client's document number.
	DocumentNumber string `json:"numero_documento"`
	// Name is the client's name.
	Name string `json:"nombre"`
	// LastName is the client's last name.
	LastName string `json:"apellido,omitempty"`
	// Email is the client's email.
	Email string `json:"email"`
	// PersonStatus is PersonNatural or PersonLegal.
	PersonStatus string `json:"personeria,omitempty"`
	// BirthDate is the birth or foundation date.
	BirthDate *Date `json:"fecha_nacimiento,omitempty"`
	// Phones lists the client's phones; at least one is required.
	Phones []Phone `json:"telefonos"`
}

func (p *Party) validate(role string) error {
	if p == nil {
		return nil
	}
	if p.DocumentNumber == "" {
		return fmt.Errorf("aex: confirm service: %s DocumentNumber is required", role)
	}
	if p.Name == "" {
		return fmt.Errorf("aex: confirm service: %s Name is required", role)
	}
	if p.Email == "" {
		return fmt.Errorf("aex: confirm service: %s Email is required", role)
	}
	if len(p.Phones) == 0 {
		return fmt.Errorf("aex: confirm service: %s requires at least one phone", role)
	}
	return nil
}

// Location is a pickup or delivery place. Set DeliveryPointID to use a
// delivery point from DeliveryPoints or RequestService (the API ignores
// the address fields in that case); otherwise fill in the address fields,
// of which AddressCode, MainStreet, CrossStreet1, and CityCode are
// required.
type Location struct {
	// DeliveryPointID is the delivery point identifier. When set, the
	// address fields are not necessary and are omitted by the API.
	DeliveryPointID FlexInt `json:"id_punto_entrega,omitempty"`
	// AddressCode uniquely identifies the address in the external
	// application.
	AddressCode string `json:"codigo,omitempty"`
	// MainStreet is the main street of the address; the API truncates it
	// to 150 characters.
	MainStreet string `json:"calle_principal,omitempty"`
	// HouseNumber is the house number.
	HouseNumber int `json:"numero_casa,omitempty"`
	// CrossStreet1 is the first cross street ("Main Street corner Cross
	// Street").
	CrossStreet1 string `json:"calle_transversal_1,omitempty"`
	// CrossStreet2 is the second cross street.
	CrossStreet2 string `json:"calle_transversal_2,omitempty"`
	// CityCode is the city code from the Cities method.
	CityCode string `json:"codigo_ciudad,omitempty"`
	// Phone is a landline phone.
	Phone int `json:"telefono,omitempty"`
	// MobilePhone is a mobile phone.
	MobilePhone int `json:"telefono_movil,omitempty"`
	// Latitude of the location.
	Latitude FlexFloat `json:"latitud,omitempty"`
	// Longitude of the location.
	Longitude FlexFloat `json:"longitud,omitempty"`
	// LocationURL points at the pickup or delivery location; an
	// alternative to coordinates.
	LocationURL string `json:"url_ubicacion,omitempty"`
	// References describes landmarks that ease locating the place; the API
	// truncates it to 250 characters.
	References string `json:"referencias,omitempty"`
	// AvailableFrom is the start of the ideal pickup or delivery time
	// slot.
	AvailableFrom *DateTime `json:"disponible_desde,omitempty"`
	// AvailableUntil is the end of the ideal pickup or delivery time slot.
	AvailableUntil *DateTime `json:"disponible_hasta,omitempty"`
	// Comment holds specific pickup or delivery instructions, e.g.,
	// lunchtime or absence days; the API truncates it to 250 characters.
	Comment string `json:"comentario,omitempty"`
	// ManagerName names the manager at the location, if not the sender.
	ManagerName string `json:"nombre_encargado,omitempty"`
	// ManagerPhone is the manager's phone, including area codes.
	ManagerPhone int `json:"telefono_encargado,omitempty"`
	// ManagerEmail is the manager's email.
	ManagerEmail string `json:"email_encargado,omitempty"`
}

func (l *Location) validate(role string) error {
	if l == nil {
		return fmt.Errorf("aex: confirm service: %s is required", role)
	}
	if l.DeliveryPointID != 0 {
		return nil
	}
	if l.AddressCode == "" {
		return fmt.Errorf("aex: confirm service: %s AddressCode is required (or set DeliveryPointID)", role)
	}
	if l.MainStreet == "" {
		return fmt.Errorf("aex: confirm service: %s MainStreet is required (or set DeliveryPointID)", role)
	}
	if l.CrossStreet1 == "" {
		return fmt.Errorf("aex: confirm service: %s CrossStreet1 is required (or set DeliveryPointID)", role)
	}
	if l.CityCode == "" {
		return fmt.Errorf("aex: confirm service: %s CityCode is required (or set DeliveryPointID)", role)
	}
	return nil
}

// NewDeliveryPointLocation returns a Location referencing the given
// delivery point identifier, for use as a pickup or delivery place.
func NewDeliveryPointLocation(deliveryPointID int) *Location {
	return &Location{DeliveryPointID: FlexInt(deliveryPointID)}
}

// ConfirmServiceParams are the parameters of the Shipments Confirm Service
// method.
type ConfirmServiceParams struct {
	authFields
	// RequestID is the service request identifier (ServiceOffer.ID)
	// returned by RequestService.
	RequestID FlexInt `json:"id_solicitud"`
	// ServiceTypeID is the selected service condition
	// (ServiceCondition.ServiceTypeID).
	ServiceTypeID FlexInt `json:"id_tipo_servicio"`
	// Sender is the sender data. When nil, the API assumes the sender from
	// the credentials used for authentication.
	Sender *Party `json:"remitente,omitempty"`
	// Pickup is the pickup data.
	Pickup *Location `json:"pickup"`
	// Recipient is the recipient data.
	Recipient *Party `json:"destinatario"`
	// Delivery is the delivery data.
	Delivery *Location `json:"entrega"`
	// AdditionalIDs lists non-mandatory additional service identifiers to
	// contract (ConditionAdditional.ID). Mandatory ones must not be sent.
	AdditionalIDs []FlexInt `json:"adicionales,omitempty"`
	// PaymentMethod is PaymentCredit (default), PaymentOrigin, or
	// PaymentDestination. When PaymentOrigin or PaymentDestination, the
	// distributor charges the recipient for the service and AEX issues the
	// invoice in the recipient's name.
	PaymentMethod string `json:"codigo_forma_pago,omitempty"`
	// TotalCollection is the final amount to collect; only required when
	// contracting the cash on delivery additional service.
	TotalCollection int `json:"total_cobro,omitempty"`
	// SerialNumber combines with GuideNumber to build a unique guide
	// identifier. Custom values must be registered with AEX
	// (soporteintegraciones@aex.com.py) in advance. Reserved values: A, C,
	// E1, FED, ML, P, RP, RR, TNT.
	SerialNumber string `json:"numero_serie,omitempty"`
	// GuideNumber combines with SerialNumber; 6 to 18 alphanumeric
	// characters.
	GuideNumber string `json:"numero_guia,omitempty"`
}

func (p ConfirmServiceParams) validate() error {
	if p.RequestID == 0 {
		return fmt.Errorf("aex: confirm service: RequestID is required")
	}
	if p.ServiceTypeID == 0 {
		return fmt.Errorf("aex: confirm service: ServiceTypeID is required")
	}
	if err := p.Pickup.validate("Pickup"); err != nil {
		return err
	}
	if p.Recipient == nil {
		return fmt.Errorf("aex: confirm service: Recipient is required")
	}
	if err := p.Recipient.validate("Recipient"); err != nil {
		return err
	}
	if err := p.Delivery.validate("Delivery"); err != nil {
		return err
	}
	if err := p.Sender.validate("Sender"); err != nil {
		return err
	}
	if (p.SerialNumber == "") != (p.GuideNumber == "") {
		return fmt.Errorf("aex: confirm service: SerialNumber and GuideNumber must be provided together")
	}
	if p.GuideNumber != "" && !guideNumberRe.MatchString(p.GuideNumber) {
		return fmt.Errorf("aex: confirm service: GuideNumber must be 6-18 alphanumeric characters (A-Z, 0-9)")
	}
	if p.SerialNumber != "" && !serialNumberRe.MatchString(p.SerialNumber) {
		return fmt.Errorf("aex: confirm service: SerialNumber must be 1-3 alphanumeric characters (A-Z, 0-9)")
	}
	return nil
}

// ConfirmService confirms an offer made by RequestService, specifying
// pickup and delivery data and the selected service condition. On success
// the transport guide is created.
func (c *Client) ConfirmService(ctx context.Context, params ConfirmServiceParams) error {
	if err := params.validate(); err != nil {
		return err
	}

	var resp baseResponse
	if err := c.doAuthenticated(ctx, confirmServicePath, &params, &resp); err != nil {
		return err
	}
	return nil
}
