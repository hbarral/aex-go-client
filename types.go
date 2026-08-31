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
