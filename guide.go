package aex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	cancelServicePath = "envios/cancelar"
	modifyGuidePath   = "actualizar-guia.php?accion=modificar"
	printGuidePath    = "envios/imprimir"

	// instructionMaxLength is the API limit for instruction texts; longer
	// values are rejected client-side.
	instructionMaxLength = 250
)

// Print formats for PrintGuide. When the format is empty or invalid, the
// API falls back to its default guide format.
const (
	// FormatLabel65x45 prints a 6.5 x 4.5 cm label with the main delivery
	// data only.
	FormatLabel65x45 = "etiqueta65x45"
	// FormatLabel8x10 prints an 8 x 10 cm label with the main delivery
	// data only.
	FormatLabel8x10 = "etiqueta8x10"
	// FormatLabel8x6 prints an 8 x 6 cm label with the main delivery data
	// only.
	FormatLabel8x6 = "etiqueta8x6"
	// FormatGuideA4 prints an A4 guide including pickup and delivery
	// instructions.
	FormatGuideA4 = "guia_A4"
)

// printGuideParams are the wire parameters of the Shipments Print method.
type printGuideParams struct {
	authFields
	// GuideNumber is the transport guide number (guia).
	GuideNumber string `json:"guia"`
	// Format indicates the format and size of the printed guide.
	Format string `json:"formato,omitempty"`
	// PrintPerItem controls product detail printing (imprimir_partida):
	// with FormatGuideA4, false prints a single guide showing up to six
	// products and true prints one guide per product; labels never show
	// product details.
	PrintPerItem bool `json:"imprimir_partida"`
}

// PrintGuide retrieves the transport guide in PDF format, returning the
// raw PDF bytes.
//
// The endpoint replies with a PDF on success, but with the standard JSON
// error envelope when the guide does not exist or is not accessible; both
// shapes are handled, the latter surfacing as *APIError.
func (c *Client) PrintGuide(ctx context.Context, guideNumber, format string, printPerItem bool) ([]byte, error) {
	if guideNumber == "" {
		return nil, fmt.Errorf("aex: print guide: guide number is required")
	}

	payload := &printGuideParams{
		GuideNumber:  guideNumber,
		Format:       format,
		PrintPerItem: printPerItem,
	}

	var pdf []byte
	err := c.runAuthenticated(ctx, printGuidePath, payload, func() error {
		body, contentType, err := c.post(ctx, printGuidePath, payload)
		if err != nil {
			return err
		}
		if !isJSONContentType(contentType) {
			pdf = body
			return nil
		}

		var resp baseResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("aex: decode response %s: %w", printGuidePath, err)
		}
		if err := checkResponse(printGuidePath, resp); err != nil {
			return err
		}
		return fmt.Errorf("aex: %s: expected PDF response, got JSON success envelope", printGuidePath)
	})
	if err != nil {
		return nil, err
	}
	return pdf, nil
}

// isJSONContentType reports whether a Content-Type header value denotes a
// JSON body.
func isJSONContentType(contentType string) bool {
	mediaType := strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	if mediaType == "" {
		return false
	}
	return strings.EqualFold(mediaType, "application/json") ||
		strings.HasSuffix(strings.ToLower(mediaType), "+json")
}

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
