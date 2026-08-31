package aex

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ResultCode is the operation result code (codigo) of the response
// envelope. The documentation defines it as a string, but the API encodes
// it inconsistently as either a JSON string ("0") or a JSON number (0);
// both shapes decode to the same value.
type ResultCode string

// UnmarshalJSON implements json.Unmarshaler, tolerating strings, numbers,
// and null.
func (r *ResultCode) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*r = ""
		return nil
	}
	*r = ResultCode(s)
	return nil
}

// baseResponse is the common envelope returned by every JSON endpoint of
// the AEX API. Concrete response types embed it.
type baseResponse struct {
	// Codigo is the operation result code. "0" means the operation was
	// processed correctly; any other value indicates an error whose
	// description is in Mensaje.
	Codigo ResultCode `json:"codigo"`
	// Mensaje contains relevant details of the executed process, or the
	// error description when Codigo is not "0".
	Mensaje string `json:"mensaje"`
}

// checkResponse validates the response envelope of an API call. It returns
// an *APIError when the result code is anything other than "0", and nil
// otherwise.
func checkResponse(endpoint string, resp baseResponse) error {
	if resp.Codigo == "" || resp.Codigo == "0" {
		return nil
	}
	return &APIError{
		Code:     string(resp.Codigo),
		Message:  resp.Mensaje,
		Endpoint: endpoint,
	}
}

// envelope makes baseResponse satisfy envelopeGetter when embedded in a
// concrete response type.
func (b baseResponse) envelope() baseResponse { return b }

// envelopeGetter is satisfied by any response type that embeds
// baseResponse, allowing shared code to inspect the response envelope.
type envelopeGetter interface {
	envelope() baseResponse
}

// responseError inspects the envelope of a decoded response and returns an
// *APIError when the result code is anything other than "0". Returns nil
// when out does not carry an envelope (or the envelope reports success).
func responseError(endpoint string, out any) error {
	if get, ok := out.(envelopeGetter); ok {
		return checkResponse(endpoint, get.envelope())
	}
	return nil
}

// credentialedRequest is satisfied by any request payload that embeds
// authFields, allowing shared code to inject the API credentials.
type credentialedRequest interface {
	setCredentials(publicKey, authorizationCode string)
}

// authFields holds the credentials that every authenticated endpoint
// requires alongside its own parameters. Request types embed it; shared
// code fills it in, so callers never set these fields themselves.
type authFields struct {
	// PublicKey (clave_publica) identifies the entity or user.
	PublicKey string `json:"clave_publica"`
	// AuthorizationCode (codigo_autorizacion) is the access token
	// obtained from the Authorization Access service.
	AuthorizationCode string `json:"codigo_autorizacion"`
}

func (a *authFields) setCredentials(publicKey, authorizationCode string) {
	a.PublicKey = publicKey
	a.AuthorizationCode = authorizationCode
}

// dateTimeLayout is the event date format used by the API:
// yyyy-mm-dd H:i:s.
const dateTimeLayout = "2006-01-02 15:04:05"

// DateTime is a time.Time that unmarshals from the API's date format
// ("2025-06-06 12:34:56"). It also tolerates RFC 3339, null, and empty
// values; the latter decode to the zero time.
type DateTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	for _, layout := range []string{dateTimeLayout, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("aex: cannot parse date %q", s)
}

// MarshalJSON implements json.Marshaler, emitting the API's date format.
// The zero time marshals as null.
func (d DateTime) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format(dateTimeLayout) + `"`), nil
}

// dateLayout is the plain date format used by the API: yyyy-mm-dd.
const dateLayout = "2006-01-02"

// Date is a time.Time that (un)marshals the API's plain date format
// ("2025-06-06"). Null and empty values decode to the zero date; the zero
// date marshals as null.
type Date struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("aex: cannot parse date %q", s)
	}
	d.Time = t
	return nil
}

// MarshalJSON implements json.Marshaler.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format(dateLayout) + `"`), nil
}

// TFBool is a boolean that the API encodes as the strings "t" and "f"
// (e.g., incluye_pickup, incluye_envio). It also tolerates true, false,
// and null on decode.
type TFBool bool

// MarshalJSON implements json.Marshaler, always emitting "t" or "f".
func (t TFBool) MarshalJSON() ([]byte, error) {
	if t {
		return []byte(`"t"`), nil
	}
	return []byte(`"f"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TFBool) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	switch s {
	case "t", "true":
		*t = true
	case "f", "false", "", "null":
		*t = false
	default:
		return fmt.Errorf("aex: cannot parse boolean %q", s)
	}
	return nil
}

// FlexInt is an integer that the API sometimes encodes as a JSON string
// (e.g., delivery point identifiers, service type identifiers). It
// unmarshals from both numbers and strings, and marshals as a number.
type FlexInt int

// MarshalJSON implements json.Marshaler.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(f))), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("aex: cannot parse integer %q", s)
	}
	*f = FlexInt(n)
	return nil
}

// FlexFloat is a float that the API sometimes encodes as a JSON string
// (e.g., costs, coordinates). It unmarshals from both numbers and
// strings, and marshals as a number.
type FlexFloat float64

// MarshalJSON implements json.Marshaler.
func (f FlexFloat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 64)), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexFloat) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("aex: cannot parse float %q", s)
	}
	*f = FlexFloat(v)
	return nil
}

// FlexBool is a boolean that the API sometimes encodes as a JSON string
// ("true"/"false") or as 1/0. It unmarshals from all of those shapes, and
// marshals as a plain boolean.
type FlexBool bool

// MarshalJSON implements json.Marshaler.
func (b FlexBool) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *FlexBool) UnmarshalJSON(data []byte) error {
	s := strings.ToLower(strings.Trim(string(data), `"`))
	switch s {
	case "true", "t", "1":
		*b = true
	case "false", "f", "0", "", "null":
		*b = false
	default:
		return fmt.Errorf("aex: cannot parse boolean %q", s)
	}
	return nil
}

// Load types for codigo_tipo_carga (P is the API default when omitted).
const (
	// LoadPackage is a conventional package (codigo_tipo_carga "P").
	LoadPackage = "P"
	// LoadDocument is a document or envelope (codigo_tipo_carga "D").
	LoadDocument = "D"
)

// Delivery point types (tipo).
const (
	// PointTypeCAC is a customer service center.
	PointTypeCAC = "CAC"
	// PointTypeELocker is a self-service terminal.
	PointTypeELocker = "ELOCKER"
)

// AdditionalService is an extra service beyond freight (e.g., insurance)
// attached to a quote or service condition.
type AdditionalService struct {
	// ID is the AEX additional service identifier (id_adicional).
	ID FlexInt `json:"id_adicional"`
	// Name is the service name (denominacion).
	Name string `json:"denominacion"`
	// Cost is the additional service cost in guaraníes.
	Cost FlexFloat `json:"costo"`
}
