// Command quickstart demonstrates the basic AEX client flow: quote a
// shipment against the sandbox environment and print the results.
//
// Set the credentials in the environment and run:
//
//	AEX_PUBLIC_KEY=... AEX_PRIVATE_KEY=... AEX_SESSION_CODE=... go run ./examples/quickstart
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hbarral/aex-go-client"
)

func main() {
	client, err := aex.New(aex.Config{
		PublicKey:   os.Getenv("AEX_PUBLIC_KEY"),
		PrivateKey:  os.Getenv("AEX_PRIVATE_KEY"),
		SessionCode: os.Getenv("AEX_SESSION_CODE"),
		Sandbox:     true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cities, err := client.Cities(ctx, aex.CitiesParams{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cities: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("coverage cities: %d\n", len(cities))
	if len(cities) >= 2 {
		fmt.Printf("quoting %s (%s) -> %s (%s)\n",
			cities[0].Name, cities[0].Code, cities[1].Name, cities[1].Code)

		quotes, err := client.Calculate(ctx, aex.CalculateParams{
			Origin:  cities[0].Code,
			Destino: cities[1].Code,
			Packages: []aex.Package{{
				Description: "Demo package",
				Weight:      1.5,
				Length:      30,
				Height:      10,
				Width:       20,
				Value:       100000,
			}},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "calculate: %v\n", err)
			os.Exit(1)
		}

		for _, quote := range quotes {
			fmt.Printf("  %-20s Gs. %-12.0f delivery in %dh (pickup: %t, delivery: %t)\n",
				quote.ServiceType, quote.FreightCost, quote.DeliveryTime,
				quote.IncludesPickup, quote.IncludesDelivery)
			for _, additional := range quote.AdditionalServices {
				fmt.Printf("    + %-20s Gs. %.0f\n", additional.Name, additional.Cost)
			}
		}
	}
}
