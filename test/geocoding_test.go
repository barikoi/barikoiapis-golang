package barikoi_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	barikoi "github.com/barikoi/barikoiapis-golang"
)

func TestReverseGeocodeSuccess(t *testing.T) {
	const respBody = `{
		"place": {
			"id": 6488,
			"distance_within_meters": 3.6856,
			"address": "House 8, Road 2, Block C, Section 2, Mirpur, Dhaka",
			"area": "Mirpur",
			"city": "Dhaka",
			"postCode": "1216",
			"address_bn": "বাড়ি ৮, রোড ২",
			"area_bn": "মিরপুর",
			"city_bn": "ঢাকা",
			"country": "বাংলাদেশ",
			"division": "ঢাকা",
			"district": "ঢাকা",
			"sub_district": "পল্লবী",
			"union": "মিরপুর",
			"pauroshova": null,
			"location_type": "শহর",
			"address_components": {"place_name": "House 8", "house": "House 8", "road": "Road 2"},
			"area_components": {"area": "Mirpur", "sub_area": "Section 2"},
			"thana": "Mirpur"
		},
		"status": 200
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got, want := r.URL.Path, "/v2/api/search/reverse/geocode"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		for param, want := range map[string]string{
			"api_key":      "test-key",
			"latitude":     "23.806703092211507",
			"longitude":    "90.35722628659195",
			"country_code": "BD",
			"country":      "true",
			"area":         "true",
		} {
			if got := queryParam(r, param); got != want {
				t.Errorf("query %s = %q, want %q", param, got, want)
			}
		}
		if _, ok := r.URL.Query()["bangla"]; ok {
			t.Error("bangla=true should not be sent when flag is false")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{
		Latitude:  23.806703092211507,
		Longitude: 90.35722628659195,
		Area:      true,
		Country:   true,
	})
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("Status = %d, want 200", resp.Status)
	}
	if got, want := resp.Place.Address, "House 8, Road 2, Block C, Section 2, Mirpur, Dhaka"; got != want {
		t.Errorf("Address = %q, want %q", got, want)
	}
	if got := resp.Place.PostCode; got != "1216" {
		t.Errorf("PostCode = %q, want \"1216\"", got)
	}
	if got := resp.Place.AddressComponents.Road; got != "Road 2" {
		t.Errorf("AddressComponents.Road = %q, want \"Road 2\"", got)
	}
	if got := resp.Place.AddressComponents.PlaceName; got != "House 8" {
		t.Errorf("AddressComponents.PlaceName = %q, want \"House 8\"", got)
	}
	if got := resp.Place.Union; got != "মিরপুর" {
		t.Errorf("Union = %q, want \"মিরপুর\"", got)
	}
}

func TestReverseGeocodeValidation(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{Latitude: 91, Longitude: 90}); !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Errorf("latitude 91: got %v, want ErrInvalidLatitude", err)
	}
	if _, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{Latitude: -90.5, Longitude: 90}); !errors.Is(err, barikoi.ErrInvalidLatitude) {
		t.Errorf("latitude -90.5: got %v, want ErrInvalidLatitude", err)
	}
	if _, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{Latitude: 23, Longitude: 181}); !errors.Is(err, barikoi.ErrInvalidLongitude) {
		t.Errorf("longitude 181: got %v, want ErrInvalidLongitude", err)
	}
}

func TestAutocompleteSuccess(t *testing.T) {
	const respBody = `{
		"places": [{
			"id": 635085,
			"longitude": 90.369999116958,
			"latitude": 23.83729875602,
			"address": "Mirpur DOHS, Mirpur DOHS",
			"city": "Dhaka",
			"area": "Dhaka University Campus",
			"district": "Dhaka",
			"postCode": 1216,
			"pType": "Admin",
			"subType": "Area",
			"uCode": "PFSU6037"
		}],
		"status": 200
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/v2/api/search/autocomplete/place"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := queryParam(r, "q"); got != "barikoi" {
			t.Errorf("q = %q, want %q", got, "barikoi")
		}
		// Bangla is a plain bool, always sent as-is.
		if got := queryParam(r, "bangla"); got != "true" {
			t.Errorf("bangla = %q, want \"true\"", got)
		}
		if _, ok := r.URL.Query()["city"]; ok {
			t.Error("city is not a documented API parameter and must not be sent")
		}
		if got := queryParam(r, "api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "barikoi", Bangla: true})
	if err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
	if len(resp.Places) != 1 {
		t.Fatalf("len(Places) = %d, want 1", len(resp.Places))
	}
	p := resp.Places[0]
	if p.ID != 635085 {
		t.Errorf("ID = %d, want 635085", p.ID)
	}
	// Coordinates are JSON numbers here but strings elsewhere; both must work.
	if p.Latitude != 23.83729875602 {
		t.Errorf("Latitude = %v, want 23.83729875602", p.Latitude)
	}
	if p.PostCode != "1216" {
		t.Errorf("PostCode = %q, want \"1216\" (number in JSON)", p.PostCode)
	}
	if p.SubType != "Area" {
		t.Errorf("SubType = %q, want \"Area\"", p.SubType)
	}
	if p.District != "Dhaka" {
		t.Errorf("District = %q, want \"Dhaka\"", p.District)
	}
}

func TestAutocompleteBanglaFalse(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := queryParam(r, "bangla"); got != "false" {
			t.Errorf("bangla = %q, want \"false\" (explicit)", got)
		}
		writeJSON(t, w, http.StatusOK, `{"places": [], "status": 200}`)
	})
	if _, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x", Bangla: false}); err != nil {
		t.Fatalf("Autocomplete: %v", err)
	}
}

func TestAutocompleteValidation(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{})
	var ve *barikoi.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("got %T, want *ValidationError", err)
	}
	if ve.Field != "q" {
		t.Errorf("Field = %q, want %q", ve.Field, "q")
	}
}

func TestGeocodeSuccess(t *testing.T) {
	const respBody = `{
		"given_address": "shawrapara",
		"fixed_address": "shewrapara, mirpur",
		"bangla_address": "শেওড়াপাড়া কবরস্থান, ইস্ট শেওড়াপাড়া",
		"address_status": "incomplete",
		"geocoded_address": {
			"Address": "Shewrapara Koborsthan, East Shewrapara",
			"address": "Shewrapara Koborsthan, East Shewrapara",
			"area": "Mirpur",
			"business_name": "Shewrapara Koborsthan",
			"city": "Dhaka",
			"district": "Dhaka",
			"sub_district": "Kafrul",
			"thana": "Kafrul",
			"pType": "Religious Place",
			"subType": "Graveyard",
			"postCode": 1216,
			"latitude": "23.79175613",
			"longitude": "90.37567053",
			"geo_location": [90.37567053, 23.79175613],
			"id": 221788,
			"uCode": "TOLJ0109",
			"place_code": "TOLJ0109"
		},
		"confidence_score_percentage": 70,
		"status": 200
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got, want := r.URL.Path, "/v2/api/search/rupantor/geocode"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := queryParam(r, "api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want form-urlencoded", ct)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		for field, want := range map[string]string{
			"q":        "shawrapara",
			"thana":    "yes",
			"district": "yes",
			"bangla":   "yes",
		} {
			if got := r.PostFormValue(field); got != want {
				t.Errorf("form %s = %q, want %q", field, got, want)
			}
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.Geocode(context.Background(), &barikoi.GeocodeRequest{Q: "shawrapara", Thana: true, District: true, Bangla: true})
	if err != nil {
		t.Fatalf("Geocode: %v", err)
	}
	if got, want := resp.FixedAddress, "shewrapara, mirpur"; got != want {
		t.Errorf("FixedAddress = %q, want %q", got, want)
	}
	if resp.AddressStatus != "incomplete" {
		t.Errorf("AddressStatus = %q, want \"incomplete\"", resp.AddressStatus)
	}
	// latitude/longitude arrive as strings; PostCode as a number.
	if resp.GeocodedAddress.Latitude != 23.79175613 {
		t.Errorf("Latitude = %v, want 23.79175613", resp.GeocodedAddress.Latitude)
	}
	if resp.GeocodedAddress.Longitude != 90.37567053 {
		t.Errorf("Longitude = %v, want 90.37567053", resp.GeocodedAddress.Longitude)
	}
	if resp.GeocodedAddress.PostCode != "1216" {
		t.Errorf("PostCode = %q, want \"1216\"", resp.GeocodedAddress.PostCode)
	}
	if resp.ConfidenceScorePercentage != 70 {
		t.Errorf("ConfidenceScorePercentage = %v, want 70", resp.ConfidenceScorePercentage)
	}
}

func TestGeocodeCapitalAddressKey(t *testing.T) {
	// Some Rupantor responses (e.g. business queries) carry the address only
	// under the capitalized "Address" key.
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{
			"given_address": "barikoi office dhaka",
			"fixed_address": "barikoi office dhaka",
			"address_status": "Incomplete",
			"confidence_score_percentage": 40,
			"status": 200,
			"geocoded_address": {
				"Address": "Barikoi HQ (barikoi.com), Dr Mohsin Plaza, House 2/7, Mirpur, Dhaka",
				"latitude": "23.823733361548",
				"longitude": "90.364025647482",
				"city": "Dhaka",
				"pType": "Office"
			}
		}`)
	})
	resp, err := c.Geocode(context.Background(), &barikoi.GeocodeRequest{Q: "barikoi office dhaka"})
	if err != nil {
		t.Fatalf("Geocode: %v", err)
	}
	if want := "Barikoi HQ (barikoi.com), Dr Mohsin Plaza, House 2/7, Mirpur, Dhaka"; resp.GeocodedAddress.Address != want {
		t.Errorf("Address = %q, want %q (from \"Address\" key)", resp.GeocodedAddress.Address, want)
	}
	if resp.GeocodedAddress.Latitude != 23.823733361548 {
		t.Errorf("Latitude = %v", resp.GeocodedAddress.Latitude)
	}
}

func TestGeocodeValidation(t *testing.T) {
	c, err := barikoi.NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Geocode(context.Background(), &barikoi.GeocodeRequest{Q: " "})
	var ve *barikoi.ValidationError
	if !errors.As(err, &ve) || ve.Field != "q" {
		t.Fatalf("got %v, want *ValidationError{Field: q}", err)
	}
}
