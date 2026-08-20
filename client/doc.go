// Package client provides a Go client for the Barikoi Location APIs
// (https://barikoi.xyz), covering geocoding, routing, and place search.
//
// Create a client with an API key obtained from https://developer.barikoi.com:
//
//	c, err := client.NewClient("YOUR_API_KEY")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := c.ReverseGeocode(context.Background(), &client.ReverseGeocodeRequest{
//		Latitude:  23.806703092211507,
//		Longitude: 90.35722628659195,
//		Area:      true,
//	})
//	if err != nil {
//		var apiErr *client.BarikoiError
//		if errors.As(err, &apiErr) && apiErr.IsAuthError() {
//			log.Fatal("invalid API key")
//		}
//		var valErr *client.ValidationError
//		if errors.As(err, &valErr) {
//			log.Fatalf("bad input field %s", valErr.Field)
//		}
//		log.Fatal(err)
//	}
//	fmt.Println(resp.Place.Address)
//
// The API key is sent as the api_key query parameter on every request, GET
// and POST alike. Use [Client.SetAPIKey] to rotate the key at runtime.
//
// Every method validates its inputs before making an HTTP call and returns
// one of three error types: [*ValidationError] for invalid input,
// [*BarikoiError] for a non-2xx API response, and [*TimeoutError] when the
// request is cancelled or times out. Distinguish them with errors.As.
package client
