package aex

import (
	"context"
	"fmt"
)

const trackingPath = "envios/tracking"

// TrackingParams are the parameters of the Shipments Tracking method.
// Exactly one of GuideNumber or OperationCode must be set; when both are
// set, the API gives priority to the guide number.
type TrackingParams struct {
	authFields
	// GuideNumber is the transport guide number to consult.
	GuideNumber string `json:"numero_guia,omitempty"`
	// OperationCode is the operation code assigned at the time of request.
	OperationCode string `json:"codigo_operacion,omitempty"`
}

func (p TrackingParams) validate() error {
	if p.GuideNumber == "" && p.OperationCode == "" {
		return fmt.Errorf("aex: tracking: GuideNumber or OperationCode is required")
	}
	return nil
}

// TrackingEvent is an event in the history of a guide, from service
// confirmation onward.
type TrackingEvent struct {
	// GuideNumber is the guide number the event belongs to. When tracking
	// by operation code, events from more than one guide may be returned.
	GuideNumber string `json:"numero_guia"`
	// Date is when the event occurred.
	Date DateTime `json:"fecha"`
	// StatusCode is the guide status code.
	StatusCode string `json:"codigo_estado"`
	// Status is the guide status description (e.g., "In route for
	// delivery").
	Status string `json:"estado"`
	// EventTypeCode is the code of the event type that occurred.
	EventTypeCode string `json:"codigo_tipo_evento"`
	// EventType is the event type description (e.g., "Received at the CDE
	// agency").
	EventType string `json:"tipo_evento"`
	// Observation holds additional details.
	Observation string `json:"observacion"`
}

// Tracking returns the history of events for a guide, ordered by event
// date from newest to oldest.
func (c *Client) Tracking(ctx context.Context, params TrackingParams) ([]TrackingEvent, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	var resp struct {
		baseResponse
		Datos []TrackingEvent `json:"datos"`
	}
	if err := c.doAuthenticated(ctx, trackingPath, &params, &resp); err != nil {
		return nil, err
	}
	return resp.Datos, nil
}

const incidentManagementPath = "envios/gestion_novedad"

// IncidentManagementParams are the parameters of the Shipments Incident
// Management method, which registers the result of an incident management.
type IncidentManagementParams struct {
	authFields
	// GuideNumber is the guide number the management refers to.
	GuideNumber string `json:"numero_guia"`
	// EventTypeCode is the code of the event type resulting from the
	// management.
	EventTypeCode string `json:"codigo_tipo_evento"`
	// ManagementDate is the date the management was carried out.
	ManagementDate *Date `json:"fecha_gestion"`
	// Observations details the management result.
	Observations string `json:"observaciones,omitempty"`
}

func (p IncidentManagementParams) validate() error {
	if p.GuideNumber == "" {
		return fmt.Errorf("aex: incident management: GuideNumber is required")
	}
	if p.EventTypeCode == "" {
		return fmt.Errorf("aex: incident management: EventTypeCode is required")
	}
	if p.ManagementDate == nil || p.ManagementDate.IsZero() {
		return fmt.Errorf("aex: incident management: ManagementDate is required")
	}
	return nil
}

// IncidentManagement registers the result of an incident management for a
// guide.
func (c *Client) IncidentManagement(ctx context.Context, params IncidentManagementParams) error {
	if err := params.validate(); err != nil {
		return err
	}

	var resp baseResponse
	if err := c.doAuthenticated(ctx, incidentManagementPath, &params, &resp); err != nil {
		return err
	}
	return nil
}
