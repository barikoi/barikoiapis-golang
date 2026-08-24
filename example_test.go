package barikoi_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/barikoi/barikoiapis-golang"
)

// Constructing a client and reverse-geocoding a coordinate. The API key comes
// from https://developer.barikoi.com.
func Example() {
	c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{
		Latitude:  23.806703092211507,
		Longitude: 90.35722628659195,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Place.Address)
}

// Geocoding an address to coordinates.
func ExampleClient_Geocode() {
	c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Geocode(context.Background(), &barikoi.GeocodeRequest{Q: "gulshan 1 circle, dhaka"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.GeocodedAddress.Latitude, resp.GeocodedAddress.Longitude)
}

// Autocomplete with Bangla script results.
func ExampleClient_Autocomplete() {
	c, err := barikoi.NewClient(os.Getenv("BARIKOI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{
		Q:      "gulshan",
		Bangla: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Places[0].AddressBn)
}

// Distinguishing the SDK's error types with errors.As. Requests are
// validated client-side before any HTTP call, so invalid input never hits
// the network.
func Example_errors() {
	c, _ := barikoi.NewClient("YOUR_API_KEY")

	_, err := c.Nearby(context.Background(), &barikoi.NearbyRequest{
		Latitude:  200, // invalid: outside [-90, 90]
		Longitude: 90.35722628659195,
		Radius:    2,
	})

	var validationErr *barikoi.ValidationError
	var apiErr *barikoi.BarikoiError
	var timeoutErr *barikoi.TimeoutError

	switch {
	case errors.As(err, &validationErr):
		fmt.Println("bad input:", validationErr)
	case errors.As(err, &apiErr):
		fmt.Println("API error:", apiErr.StatusCode)
	case errors.As(err, &timeoutErr):
		fmt.Println("timed out:", timeoutErr)
	}
}
