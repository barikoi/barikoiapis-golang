package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestSearchPlaceSuccess(t *testing.T) {
	const respBody = `{
		"places": [{
			"address": "Barikoi HQ (barikoi.com), Dr Mohsin Plaza, House 2/7, Begum Rokeya Sarani, Pallabi, Mirpur, Dhaka",
			"place_code": "BKOI2017"
		}],
		"session_id": "eec76e23-481f-4060-8959-df5db6219126",
		"status": 200
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/v2/search-place"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := r.URL.Query().Get("q"); got != "barikoi" {
			t.Errorf("q = %q, want %q", got, "barikoi")
		}
		if got := r.URL.Query().Get("api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.SearchPlace(context.Background(), &SearchPlaceRequest{Q: "barikoi"})
	if err != nil {
		t.Fatalf("SearchPlace: %v", err)
	}
	if len(resp.Places) != 1 || resp.Places[0].PlaceCode != "BKOI2017" {
		t.Fatalf("places = %+v, want one with place_code BKOI2017", resp.Places)
	}
	if resp.SessionID != "eec76e23-481f-4060-8959-df5db6219126" {
		t.Errorf("session_id = %q, want eec76e23-...", resp.SessionID)
	}
}

func TestSearchPlaceValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SearchPlace(context.Background(), &SearchPlaceRequest{})
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "q" {
		t.Fatalf("got %v, want *ValidationError{Field: q}", err)
	}
}

func TestPlaceDetailsSuccess(t *testing.T) {
	const respBody = `{
		"session_id": "8d6a20cb-e07d-4332-a293-d0cf0fce968e",
		"status": 200,
		"place": {
			"address": "Barikoi HQ (barikoi.com), Dr Mohsin Plaza, House 2/7, Begum Rokeya Sarani, Pallabi, Mirpur, Dhaka",
			"place_code": "BKOI2017",
			"latitude": "23.823730671721",
			"longitude": "90.36402004477634"
		}
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/v2/places"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		q := r.URL.Query()
		if got := q.Get("place_code"); got != "BKOI2017" {
			t.Errorf("place_code = %q, want %q", got, "BKOI2017")
		}
		if got := q.Get("session_id"); got != "8d6a20cb-e07d-4332-a293-d0cf0fce968e" {
			t.Errorf("session_id = %q, want 8d6a20cb-...", got)
		}
		if got := q.Get("api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.PlaceDetails(context.Background(), &PlaceDetailsRequest{
		PlaceCode: "BKOI2017",
		SessionID: "8d6a20cb-e07d-4332-a293-d0cf0fce968e",
	})
	if err != nil {
		t.Fatalf("PlaceDetails: %v", err)
	}
	// Coordinates arrive as strings; FlexFloat must handle them.
	if resp.Place.Latitude != 23.823730671721 {
		t.Errorf("latitude = %v, want 23.823730671721", resp.Place.Latitude)
	}
	if resp.Place.Longitude != 90.36402004477634 {
		t.Errorf("longitude = %v, want 90.36402004477634", resp.Place.Longitude)
	}
}

func TestPlaceDetailsValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	var ve *ValidationError
	_, err = c.PlaceDetails(context.Background(), &PlaceDetailsRequest{SessionID: "s"})
	if !errors.As(err, &ve) || ve.Field != "place_code" {
		t.Fatalf("got %v, want *ValidationError{Field: place_code}", err)
	}
	_, err = c.PlaceDetails(context.Background(), &PlaceDetailsRequest{PlaceCode: "BKOI2017"})
	if !errors.As(err, &ve) || ve.Field != "session_id" {
		t.Fatalf("got %v, want *ValidationError{Field: session_id}", err)
	}
}

func TestNearbySuccess(t *testing.T) {
	const respBody = `{
		"places": [{
			"id": "6488",
			"name": "Uttara sector 12 Park",
			"distance_in_meters": 0.00045241879869987946,
			"longitude": 90.383051633835,
			"latitude": 23.871887192063,
			"type": "Recreation",
			"address": "Road 11, Sector 12",
			"area": "Uttara",
			"city": "Dhaka",
			"postcode": "1230",
			"sub_type": "PARK",
			"place_code": "OYFR1074"
		}],
		"status": 200
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/v2/api/search/nearby/0.5/10"; got != want {
			t.Errorf("path = %q, want %q (radius/limit as path parameters)", got, want)
		}
		q := r.URL.Query()
		if got := q.Get("latitude"); got != "23.871887192063" {
			t.Errorf("latitude = %q, want %q", got, "23.871887192063")
		}
		if got := q.Get("longitude"); got != "90.383051633835" {
			t.Errorf("longitude = %q, want %q", got, "90.383051633835")
		}
		if got := q.Get("api_key"); got != "test-key" {
			t.Errorf("api_key = %q, want %q", got, "test-key")
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.Nearby(context.Background(), &NearbyRequest{
		Latitude:  23.871887192063,
		Longitude: 90.383051633835,
		Radius:    0.5,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("Nearby: %v", err)
	}
	if len(resp.Places) != 1 || resp.Places[0].Name != "Uttara sector 12 Park" {
		t.Fatalf("places = %+v", resp.Places)
	}
	if resp.Places[0].PostCode != "1230" {
		t.Errorf("postCode = %q, want \"1230\"", resp.Places[0].PostCode)
	}
	if resp.Places[0].ID != "6488" {
		t.Errorf("id = %q, want \"6488\" (string)", resp.Places[0].ID)
	}
	if resp.Places[0].PType != "Recreation" {
		t.Errorf("type = %q, want Recreation", resp.Places[0].PType)
	}
	if resp.Places[0].SubType != "PARK" {
		t.Errorf("sub_type = %q, want PARK", resp.Places[0].SubType)
	}
	if resp.Places[0].PlaceCode != "OYFR1074" {
		t.Errorf("place_code = %q, want OYFR1074", resp.Places[0].PlaceCode)
	}
	if resp.Places[0].DistanceInMeters != 0.00045241879869987946 {
		t.Errorf("distance_in_meters = %v", resp.Places[0].DistanceInMeters)
	}
}

func TestNearbyValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	var ve *ValidationError
	cases := []struct {
		name      string
		req       *NearbyRequest
		wantField string
	}{
		{"radius too small", &NearbyRequest{Latitude: 23.8, Longitude: 90.3, Radius: 0.05, Limit: 10}, "radius"},
		{"radius too large", &NearbyRequest{Latitude: 23.8, Longitude: 90.3, Radius: 101, Limit: 10}, "radius"},
		{"limit too small", &NearbyRequest{Latitude: 23.8, Longitude: 90.3, Radius: 1, Limit: 0}, "limit"},
		{"limit too large", &NearbyRequest{Latitude: 23.8, Longitude: 90.3, Radius: 1, Limit: 101}, "limit"},
		{"bad latitude", &NearbyRequest{Latitude: 95, Longitude: 90.3, Radius: 1, Limit: 10}, ""},
	}
	for _, tc := range cases {
		_, err := c.Nearby(context.Background(), tc.req)
		if err == nil {
			t.Errorf("%s: want error, got nil", tc.name)
			continue
		}
		if tc.wantField == "" {
			if !errors.Is(err, ErrInvalidLatitude) {
				t.Errorf("%s: got %v, want ErrInvalidLatitude", tc.name, err)
			}
			continue
		}
		if !errors.As(err, &ve) || ve.Field != tc.wantField {
			t.Errorf("%s: got %v, want *ValidationError{Field: %s}", tc.name, err, tc.wantField)
		}
	}
}

func TestCheckNearbySuccess(t *testing.T) {
	const respBody = `{
		"message": "Inside geo fence",
		"status": 200,
		"data": {
			"id": "68e5f2ab382b2",
			"name": "destination",
			"radius": "50",
			"latitude": "23.76245538673939",
			"longitude": "90.37852866512583",
			"user_id": 1624
		}
	}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/v2/api/check/nearby"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		q := r.URL.Query()
		for param, want := range map[string]string{
			"api_key":               "test-key",
			"current_latitude":      "23.762412943322726",
			"current_longitude":     "90.37864864706823",
			"destination_latitude":  "23.76245538673939",
			"destination_longitude": "90.37852866512583",
			"radius":                "50",
		} {
			if got := q.Get(param); got != want {
				t.Errorf("query %s = %q, want %q", param, got, want)
			}
		}
		writeJSON(t, w, http.StatusOK, respBody)
	})

	resp, err := c.CheckNearby(context.Background(), &CheckNearbyRequest{
		CurrentLatitude:      23.762412943322726,
		CurrentLongitude:     90.37864864706823,
		DestinationLatitude:  23.76245538673939,
		DestinationLongitude: 90.37852866512583,
		Radius:               50,
	})
	if err != nil {
		t.Fatalf("CheckNearby: %v", err)
	}
	if resp.Message != "Inside geo fence" {
		t.Errorf("message = %q, want \"Inside geo fence\"", resp.Message)
	}
	if resp.Data == nil || resp.Data.Name != "destination" || resp.Data.UserID != 1624 {
		t.Errorf("data = %+v, want name \"destination\", user_id 1624", resp.Data)
	}
}

func TestCheckNearbyOutsideRadius(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"message": "Point not inside polygon", "status": 200, "data": null}`)
	})
	resp, err := c.CheckNearby(context.Background(), &CheckNearbyRequest{
		CurrentLatitude:      23.76,
		CurrentLongitude:     90.37,
		DestinationLatitude:  23.77,
		DestinationLongitude: 90.38,
		Radius:               50,
	})
	if err != nil {
		t.Fatalf("CheckNearby: %v", err)
	}
	if resp.Data != nil {
		t.Errorf("data = %+v, want nil", resp.Data)
	}
}

func TestCheckNearbyValidation(t *testing.T) {
	c, err := NewClient("k")
	if err != nil {
		t.Fatal(err)
	}
	var ve *ValidationError
	_, err = c.CheckNearby(context.Background(), &CheckNearbyRequest{
		CurrentLatitude:      23.7,
		CurrentLongitude:     90.3,
		DestinationLatitude:  23.8,
		DestinationLongitude: 90.4,
		Radius:               5,
	})
	if !errors.As(err, &ve) || ve.Field != "radius" {
		t.Fatalf("radius 5: got %v, want *ValidationError{Field: radius}", err)
	}
	_, err = c.CheckNearby(context.Background(), &CheckNearbyRequest{
		CurrentLatitude:      23.7,
		CurrentLongitude:     90.3,
		DestinationLatitude:  23.8,
		DestinationLongitude: 90.4,
		Radius:               1001,
	})
	if !errors.As(err, &ve) || ve.Field != "radius" {
		t.Fatalf("radius 1001: got %v, want *ValidationError{Field: radius}", err)
	}
	_, err = c.CheckNearby(context.Background(), &CheckNearbyRequest{
		CurrentLatitude:      23.7,
		CurrentLongitude:     90.3,
		DestinationLatitude:  95,
		DestinationLongitude: 90.4,
		Radius:               50,
	})
	if !errors.Is(err, ErrInvalidLatitude) {
		t.Errorf("bad destination latitude: got %v, want ErrInvalidLatitude", err)
	}
}
