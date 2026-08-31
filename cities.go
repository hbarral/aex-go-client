package aex

import "context"

const citiesPath = "envios/ciudades"

// CitiesParams are the parameters of the Shipments Cities method.
type CitiesParams struct {
	authFields
	// Origin optionally filters the result to cities enabled as
	// destinations from the given origin city code.
	Origin string `json:"origen,omitempty"`
}

// City is a city with shipping coverage.
type City struct {
	// Code is the city code (codigo_ciudad), used for service requests.
	Code string `json:"codigo_ciudad"`
	// Name is the city description.
	Name string `json:"denominacion"`
	// DepartmentCode is the code of the department the city belongs to.
	DepartmentCode string `json:"codigo_departamento"`
	// DepartmentName is the department/state name.
	DepartmentName string `json:"departamento_denominacion"`
	// CountryCode is the code of the country the city belongs to.
	CountryCode string `json:"codigo_pais"`
	// CountryName is the country name.
	CountryName string `json:"pais_denominacion"`
}

// Cities returns the list of cities with coverage, optionally filtered by
// origin city.
func (c *Client) Cities(ctx context.Context, params CitiesParams) ([]City, error) {
	var resp struct {
		baseResponse
		Datos []City `json:"datos"`
	}
	if err := c.doAuthenticated(ctx, citiesPath, &params, &resp); err != nil {
		return nil, err
	}
	return resp.Datos, nil
}
