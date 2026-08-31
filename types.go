package aex

// baseResponse is the common envelope returned by every JSON endpoint of
// the AEX API. Concrete response types embed it.
type baseResponse struct {
	// Codigo is the operation result code. "0" means the operation was
	// processed correctly; any other value indicates an error whose
	// description is in Mensaje.
	Codigo string `json:"codigo"`
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
		Code:     resp.Codigo,
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
