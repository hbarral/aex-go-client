# aex-go-client

Go client for the AEX Paraguay courier and shipping API (v1.5.4). Covers
authorization, city and delivery point lookups, shipping quotes, service
requests and confirmations, guide printing and updates, tracking,
inventory, and webhook handling.

Designed to be embedded in a Go web application: stdlib only, no runtime
dependencies, safe for concurrent use, context-aware.

## Installation

```sh
go get github.com/hbarral/aex-go-client
```

## Usage

### Client setup

```go
client, err := aex.New(aex.Config{
	PublicKey:   "your-public-key",
	PrivateKey:  "your-private-key",
	SessionCode: "your-session-code",
	Sandbox:     true, // https://sandbox.aex.com.py; false targets production
})
```

The session code is a string **chosen by your application**, not a
credential issued by AEX: it can be a fixed value (e.g., `"prueba"`) or a
dynamic one generated per session (e.g., a UUID). Whatever value you pick
is sent in plain text alongside the hash so the API can validate it.

The private key is only ever sent as `md5(privateKey + sessionCode)`, as
the API requires. Authorization codes are obtained automatically, cached,
and refreshed 2 minutes before their 10-minute expiry; if a cached code is
rejected server-side, the client re-authenticates and retries the request
once. Call `client.Authenticate(ctx)` to validate credentials up front;
failures wrap `aex.ErrUnauthorized`.

### Quoting a shipment

```go
quotes, err := client.Calculate(ctx, aex.CalculateParams{
	Origin:  "ASU",
	Destino: "CDE",
	Packages: []aex.Package{{
		Description: "Camiseta",
		Weight:      1.5,
		Length:      30,
		Height:      10,
		Width:       20,
		Value:       100000, // declared value in PYG
	}},
})
```

### Booking a service (request → confirm → cancel)

```go
offer, err := client.RequestService(ctx, aex.RequestServiceParams{
	Origin:        "ASU",
	Destino:       "CDE",
	OperationCode: "PEDIDO-12345",
	Packages:      packages,
})

err = client.ConfirmService(ctx, aex.ConfirmServiceParams{
	RequestID:     offer.ID,
	ServiceTypeID: offer.Conditions[0].ServiceTypeID,
	Pickup: &aex.Location{ /* address fields, or: */ },
	Delivery:      aex.NewDeliveryPointLocation(17),
	Recipient: &aex.Party{
		DocumentNumber: "1234567",
		Name:           "Juan",
		Email:          "juan@example.com",
		Phones:         []aex.Phone{{Number: 98111222}},
	},
})

err = client.CancelService(ctx, "A002866303")
```

`Location` accepts either a delivery point ID or a full address (code,
main street, cross street, city code are required); `Party` requires
document number, name, email, and at least one phone. All of this is
validated client-side before a request is sent.

### Tracking, guide updates, and printing

```go
events, err := client.Tracking(ctx, aex.TrackingParams{GuideNumber: "A002866303"})
// or: aex.TrackingParams{OperationCode: "PEDIDO-12345"}

err = client.ModifyGuide(ctx, aex.ModifyGuideParams{
	GuideNumber: "A002866303",
	Modify:      aex.GuideModifications{DeliveryInstructions: "Entregar en recepcion"},
})

pdf, err := client.PrintGuide(ctx, "A002866303", aex.FormatLabel8x6, false)
```

### Inventory (for AEX-stored products)

```go
stock, err := client.Inventory(ctx, aex.InventoryParams{})           // all products
stock, err = client.Inventory(ctx, aex.InventoryParams{ProductCodes: []string{"SKU-1"}})
```

### Receiving webhooks

```go
http.Handle("/webhooks/aex", aex.HandleWebhook(
	func(ctx context.Context, event aex.WebhookEvent) error {
		// persist the event; return an error to make AEX retry
		return nil
	},
	aex.WithWebhookAuth("Authorization", "Bearer your-token"),
))
```

The handler replies with the acknowledgement shape AEX expects: HTTP 200
`{"isSuccess":true}` on success, HTTP 400 `{"isSuccess":false}` otherwise
(AEX retries failed notifications up to 4 times). Use `aex.ParseWebhook`
if you route requests yourself.

### Error handling

All API-level failures surface as `*aex.APIError` (result code and
message), transport failures as `*aex.HTTPError`, and rejected
credentials as errors wrapping `aex.ErrUnauthorized`:

```go
var apiErr *aex.APIError
if errors.As(err, &apiErr) {
	log.Printf("AEX error %s: %s", apiErr.Code, apiErr.Message)
}
```

### Data type notes

The API encodes values inconsistently (documented numbers arrive as JSON
strings and vice versa, confirmed on the sandbox); the client normalizes
all of it:

- Result codes (`codigo`) decode from numbers or strings (`ResultCode`).
- Numeric fields (identifiers, costs, quantities, coordinates) decode from
  numbers or strings into `aex.FlexInt` / `aex.FlexFloat`.
- Booleans decode from `true`/`false`, `"true"`/`"false"`, and `1`/`0`
  into `aex.FlexBool`; the `"t"`/`"f"` flags (incluye_pickup, etc.) into
  `aex.TFBool`.
- Dates parse from `yyyy-mm-dd H:i:s` (`aex.DateTime`) and `yyyy-mm-dd`
  (`aex.Date`), with null/empty tolerance.

## Testing

```sh
go test ./...                 # unit tests (httptest, no network)
go test -race ./...           # with the race detector
go vet ./...
gofmt -l .
```

Optional sandbox smoke tests (require credentials):

```sh
AEX_PUBLIC_KEY=... AEX_PRIVATE_KEY=... AEX_SESSION_CODE=... \
	go test -tags integration -run Integration -v
```

A runnable example lives in [examples/quickstart](examples/quickstart).

## Reference

- [Official AEX API documentation](https://www.aex.com.py/web/documentacion-api.php)
