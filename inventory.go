package aex

import "context"

const inventoryPath = "inventario/existencia"

// InventoryParams are the parameters of the Inventory Existence method.
type InventoryParams struct {
	authFields
	// ProductCodes optionally restricts the query to the given product
	// codes. When empty, the API returns the complete list of stored
	// products for the account.
	ProductCodes []string `json:"codigos_producto,omitempty"`
}

// ProductStock is the available quantity of a product stored by AEX.
type ProductStock struct {
	// ProductCode is the code agreed between parties to identify the
	// product in the warehouse.
	ProductCode string `json:"codigo_producto"`
	// Stock is the available product quantity in inventory.
	Stock int `json:"existencia"`
	// Name is the product name.
	Name string `json:"denominacion"`
}

// Inventory returns the list of stored products with their available
// quantities, optionally restricted to the given product codes.
func (c *Client) Inventory(ctx context.Context, params InventoryParams) ([]ProductStock, error) {
	var resp struct {
		baseResponse
		Datos []ProductStock `json:"datos"`
	}
	if err := c.doAuthenticated(ctx, inventoryPath, &params, &resp); err != nil {
		return nil, err
	}
	return resp.Datos, nil
}
