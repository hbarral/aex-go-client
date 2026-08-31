package aex

import (
	"context"
	"fmt"
)

const (
	cancelServicePath = "envios/cancelar"
	modifyGuidePath   = "actualizar-guia.php?accion=modificar"

	// instructionMaxLength is the API limit for instruction texts; longer
	// values are rejected client-side.
	instructionMaxLength = 250
)

// CancelService cancels the service of the given transport guide number.
// Only services that have not had operational management can be
// cancelled.
func (c *Client) CancelService(ctx context.Context, guideNumber string) error {
	if guideNumber == "" {
		return fmt.Errorf("aex: cancel service: guide number is required")
	}

	var payload struct {
		authFields
		GuideNumber string `json:"numero_guia"`
	}
	payload.GuideNumber = guideNumber

	var resp baseResponse
	if err := c.doAuthenticated(ctx, cancelServicePath, &payload, &resp); err != nil {
		return err
	}
	return nil
}

// GuideModifications holds the guide fields that can be modified through
// ModifyGuide. At least one instruction must be set; empty strings are
// rejected by the API, so only set fields are sent.
type GuideModifications struct {
	// PickupInstructions is the instructions text for pickup.
	PickupInstructions string `json:"indicacion_pickup,omitempty"`
	// DeliveryInstructions is the instructions text for delivery.
	DeliveryInstructions string `json:"indicacion_entrega,omitempty"`
}

func (m GuideModifications) validate() error {
	if m.PickupInstructions == "" && m.DeliveryInstructions == "" {
		return fmt.Errorf("aex: modify guide: at least one instruction is required")
	}
	if len(m.PickupInstructions) > instructionMaxLength {
		return fmt.Errorf("aex: modify guide: PickupInstructions exceeds %d characters", instructionMaxLength)
	}
	if len(m.DeliveryInstructions) > instructionMaxLength {
		return fmt.Errorf("aex: modify guide: DeliveryInstructions exceeds %d characters", instructionMaxLength)
	}
	return nil
}

// ModifyGuideParams are the parameters of the Update Guide Modify method.
type ModifyGuideParams struct {
	authFields
	// GuideNumber is the complete guide number to modify (e.g.,
	// A002866303).
	GuideNumber string `json:"numero_guia"`
	// Modify holds the fields to modify; they must be nested here, not
	// sent at the top level.
	Modify GuideModifications `json:"modificar"`
}

func (p ModifyGuideParams) validate() error {
	if p.GuideNumber == "" {
		return fmt.Errorf("aex: modify guide: GuideNumber is required")
	}
	return p.Modify.validate()
}

// ModifyGuide modifies the enabled data of a guide belonging to the
// authenticated client: currently pickup and/or delivery instructions.
// Absent fields are not modified; empty or disallowed fields generate an
// API error.
func (c *Client) ModifyGuide(ctx context.Context, params ModifyGuideParams) error {
	if err := params.validate(); err != nil {
		return err
	}

	var resp baseResponse
	if err := c.doAuthenticated(ctx, modifyGuidePath, &params, &resp); err != nil {
		return err
	}
	return nil
}
