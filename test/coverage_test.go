package barikoi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	barikoi "github.com/barikoi/barikoiapis-golang"
	"github.com/barikoi/barikoiapis-golang/client"
)

// timeoutErr is a net.Error whose Timeout reports true.
type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

// TestWrapTransportErrorNetTimeout covers the net.Error branch of
// wrapTransportError (a raw transport timeout that is not a context error).
func TestWrapTransportErrorNetTimeout(t *testing.T) {
	// Indirect through the client package so the test asserts observable
	// behavior: a custom RoundTripper whose error is a net.Error timeout is
	// surfaced as *TimeoutError.
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, timeoutErr{}
	})
	c, err := barikoi.NewClient("k", barikoi.WithAllowInsecure(),
		barikoi.WithBaseURL("http://127.0.0.1:1"),
		barikoi.WithHTTPClient(&http.Client{Transport: rt}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.SnapToRoad(context.Background(), &barikoi.SnapToRoadRequest{Point: "23.8,90.4"})
	if err == nil {
		t.Fatal("SnapToRoad: expected error, got nil")
	}
	var te *barikoi.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("error = %T (%v), want *TimeoutError", err, err)
	}
}

// TestErrorCodeForStatusRemaining covers the codeForStatus branches not
// exercised elsewhere: 400, 402, and the unknown default.
func TestErrorCodeForStatusRemaining(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{http.StatusBadRequest, "missing_parameter"},
		{http.StatusPaymentRequired, "payment_exception"},
		{http.StatusTeapot, "unknown_error"},
	}
	for _, tc := range cases {
		c := offlineClient(t, tc.status, `{"message": "x"}`)
		_, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"})
		var be *barikoi.BarikoiError
		if !errors.As(err, &be) {
			t.Fatalf("status %d: error = %T, want *BarikoiError", tc.status, err)
		}
		if be.Code != tc.want {
			t.Errorf("status %d: Code = %q, want %q", tc.status, be.Code, tc.want)
		}
	}
}

// TestDecodeError covers the JSON decoding failure branch of do.
func TestDecodeError(t *testing.T) {
	c := offlineClient(t, http.StatusOK, `{"places": nope}`)
	_, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"})
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

// TestInvalidFloatString covers FlexFloat receiving a non-numeric string.
func TestInvalidFloatString(t *testing.T) {
	var f client.FlexFloat
	if err := json.Unmarshal([]byte(`"not-a-number"`), &f); err == nil {
		t.Fatal("expected error for non-numeric string")
	}
}

// TestNilShortCircuit ensures a nil-typed slice round-trips through the
// FlexFloat/FlexString custom unmarshalers' number branches.
func TestFlexNumberBranches(t *testing.T) {
	var s client.FlexString
	if err := json.Unmarshal([]byte(`123`), &s); err != nil {
		t.Fatalf("FlexString number: %v", err)
	}
	if s != "123" {
		t.Errorf("FlexString = %q, want 123", s)
	}
}

// TestGeoJSONUnknownShape covers RouteGeometry's raw-JSON fallback: an
// object without a type field lands in Polyline.
func TestGeoJSONUnknownShape(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"code":"Ok","routes":[{"geometry":{"weird":true},"distance":1,"duration":1}],"waypoints":[]}`)
	})
	resp, err := c.RouteOverview(context.Background(), &barikoi.RouteOverviewRequest{
		Coordinates: "90.3,23.8;90.4,23.9",
	})
	if err != nil {
		t.Fatalf("RouteOverview: %v", err)
	}
	g := resp.Routes[0].Geometry
	if g.GeoJSON != nil || g.Polyline != `{"weird":true}` {
		t.Errorf("Geometry = %+v, want raw JSON in Polyline", g)
	}
}

// TestPlaceListUnmarshalError covers the geocoded place alias unmarshal
// failure when the geocoded_address payload is not an object.
func TestPlaceListUnmarshalError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"given_address":"x","geocoded_address":42,"status":200}`)
	})
	_, err := c.Geocode(context.Background(), &barikoi.GeocodeRequest{Q: "x"})
	if err == nil {
		t.Fatal("expected unmarshal error, got nil")
	}
}

// TestContextCancellationBeforeCall covers the ctx.Err() branch of
// wrapTransportError: a context cancelled while the request is in flight.
func TestContextCancellationBeforeCall(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		writeJSON(t, w, http.StatusOK, `{"places":[],"status":200}`)
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	_, err := c.SearchPlace(ctx, &barikoi.SearchPlaceRequest{Q: "x"})
	var te *barikoi.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("error = %T (%v), want *TimeoutError", err, err)
	}
}

// TestPerRequestTimeoutApplies ensures SetTimeout's value is used on the
// next request (short timeout -> TimeoutError against a slow stub).
func TestPerRequestTimeoutApplies(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		writeJSON(t, w, http.StatusOK, `{"places":[],"status":200}`)
	})
	c.SetTimeout(50 * time.Millisecond)
	_, err := c.Autocomplete(context.Background(), &barikoi.AutocompleteRequest{Q: "x"})
	var te *barikoi.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("error = %T (%v), want *TimeoutError", err, err)
	}
}

// TestTransportErrorPassthrough covers wrapTransportError's final branch:
// a non-timeout transport error is returned unchanged.
func TestTransportErrorPassthrough(t *testing.T) {
	boom := errors.New("connection refused")
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, boom })
	c, err := barikoi.NewClient("k", barikoi.WithAllowInsecure(),
		barikoi.WithBaseURL("http://127.0.0.1:1"),
		barikoi.WithHTTPClient(&http.Client{Transport: rt}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.SearchPlace(context.Background(), &barikoi.SearchPlaceRequest{Q: "x"})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want the original transport error", err)
	}
}

// TestRemainingMethodTransportErrors covers the do() error-return branch
// of every method wrapper that had no failing-request test of its own.
func TestRemainingMethodTransportErrors(t *testing.T) {
	boom := errors.New("connection refused")
	newClient := func(t *testing.T) *barikoi.Client {
		t.Helper()
		rt := roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, boom })
		c, err := barikoi.NewClient("k", barikoi.WithAllowInsecure(),
			barikoi.WithBaseURL("http://127.0.0.1:1"),
			barikoi.WithHTTPClient(&http.Client{Transport: rt}))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		return c
	}
	ctx := context.Background()
	pts := []barikoi.OptimizeRoutePoint{{ID: 1, Point: "23.8,90.4"}}
	calls := []struct {
		name string
		call func() error
	}{
		{"ReverseGeocode", func() error {
			_, e := newClient(t).ReverseGeocode(ctx, &barikoi.ReverseGeocodeRequest{Latitude: 23.8, Longitude: 90.4})
			return e
		}},
		{"PlaceDetails", func() error {
			_, e := newClient(t).PlaceDetails(ctx, &barikoi.PlaceDetailsRequest{PlaceCode: "PXW", SessionID: "eec76e23-481f-4060-8959-df5db6219126"})
			return e
		}},
		{"Nearby", func() error {
			_, e := newClient(t).Nearby(ctx, &barikoi.NearbyRequest{Latitude: 23.8, Longitude: 90.4})
			return e
		}},
		{"CheckNearby", func() error {
			_, e := newClient(t).CheckNearby(ctx, &barikoi.CheckNearbyRequest{Radius: 100})
			return e
		}},
		{"RouteOverview", func() error {
			_, e := newClient(t).RouteOverview(ctx, &barikoi.RouteOverviewRequest{Coordinates: "90.3,23.8;90.4,23.9"})
			return e
		}},
		{"CalculateRoute", func() error { _, e := newClient(t).CalculateRoute(ctx, &barikoi.CalculateRouteRequest{}); return e }},
		{"OptimizeRoute", func() error {
			_, e := newClient(t).OptimizeRoute(ctx, &barikoi.OptimizeRouteRequest{Source: "23.8,90.4", Destination: "23.7,90.3", GeoPoints: pts})
			return e
		}},
	}
	for _, tc := range calls {
		if err := tc.call(); !errors.Is(err, boom) {
			t.Errorf("%s: error = %v, want the original transport error", tc.name, err)
		}
	}
}

// TestFlexBadTypes covers the custom unmarshalers' rejection branches:
// neither valid JSON numbers nor numeric strings.
func TestFlexBadTypes(t *testing.T) {
	var f client.FlexFloat
	if err := json.Unmarshal([]byte(`true`), &f); err == nil {
		t.Error("FlexFloat(true): expected error")
	}
	var s client.FlexString
	if err := json.Unmarshal([]byte(`true`), &s); err == nil {
		t.Error("FlexString(true): expected error")
	}
}

// errBody fails on Read to cover the do() response-body read error path.
type errBody struct{}

func (errBody) Read([]byte) (int, error) { return 0, timeoutErr{} }
func (errBody) Close() error             { return nil }

// TestResponseBodyReadError covers the io.ReadAll failure branch of do.
func TestResponseBodyReadError(t *testing.T) {
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: errBody{}, Header: http.Header{}}, nil
	})
	c, err := barikoi.NewClient("k", barikoi.WithAllowInsecure(),
		barikoi.WithBaseURL("http://127.0.0.1:1"),
		barikoi.WithHTTPClient(&http.Client{Transport: rt}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.SearchPlace(context.Background(), &barikoi.SearchPlaceRequest{Q: "x"})
	if err == nil {
		t.Fatal("expected read error, got nil")
	}
}

// TestValidationShortCircuitPerMethod covers the remaining early-return
// validation branches per method (second coord invalid, etc.).
func TestValidationShortCircuitPerMethod(t *testing.T) {
	ctx := context.Background()
	c := offlineClient(t, http.StatusOK, `{}`) // never reached

	cases := []struct {
		name string
		call func() error
	}{
		{"reverseGeocode bad lat", func() error {
			_, err := c.ReverseGeocode(ctx, &barikoi.ReverseGeocodeRequest{Latitude: 91, Longitude: 0})
			return err
		}},
		{"geocode blank q", func() error {
			_, err := c.Geocode(ctx, &barikoi.GeocodeRequest{Q: " "})
			return err
		}},
		{"searchPlace blank q", func() error {
			_, err := c.SearchPlace(ctx, &barikoi.SearchPlaceRequest{Q: ""})
			return err
		}},
		{"placeDetails blank code", func() error {
			_, err := c.PlaceDetails(ctx, &barikoi.PlaceDetailsRequest{})
			return err
		}},
		{"placeDetails bad uuid", func() error {
			_, err := c.PlaceDetails(ctx, &barikoi.PlaceDetailsRequest{PlaceCode: "PXW", SessionID: "not-uuid"})
			return err
		}},
		{"nearby bad lat", func() error {
			_, err := c.Nearby(ctx, &barikoi.NearbyRequest{Latitude: -91, Longitude: 0})
			return err
		}},
		{"checkNearby bad current", func() error {
			_, err := c.CheckNearby(ctx, &barikoi.CheckNearbyRequest{CurrentLatitude: 200, CurrentLongitude: 0, DestinationLatitude: 0, DestinationLongitude: 0, Radius: 100})
			return err
		}},
		{"checkNearby bad destination", func() error {
			_, err := c.CheckNearby(ctx, &barikoi.CheckNearbyRequest{CurrentLatitude: 0, CurrentLongitude: 0, DestinationLatitude: 0, DestinationLongitude: 181, Radius: 100})
			return err
		}},
		{"routeOverview bad coords", func() error {
			_, err := c.RouteOverview(ctx, &barikoi.RouteOverviewRequest{Coordinates: "one;two"})
			return err
		}},
		{"calculateRoute bad start", func() error {
			_, err := c.CalculateRoute(ctx, &barikoi.CalculateRouteRequest{Start: barikoi.Coordinate{Latitude: 91, Longitude: 0}})
			return err
		}},
		{"calculateRoute bad destination", func() error {
			_, err := c.CalculateRoute(ctx, &barikoi.CalculateRouteRequest{Start: barikoi.Coordinate{}, Destination: barikoi.Coordinate{Longitude: -181}})
			return err
		}},
		{"optimizeRoute bad source", func() error {
			_, err := c.OptimizeRoute(ctx, &barikoi.OptimizeRouteRequest{Source: "x", Destination: "23.8,90.4", GeoPoints: []barikoi.OptimizeRoutePoint{{ID: 1, Point: "23.8,90.4"}}})
			return err
		}},
		{"optimizeRoute bad destination", func() error {
			_, err := c.OptimizeRoute(ctx, &barikoi.OptimizeRouteRequest{Source: "23.8,90.4", Destination: "x", GeoPoints: []barikoi.OptimizeRoutePoint{{ID: 1, Point: "23.8,90.4"}}})
			return err
		}},
		{"snapToRoad bad point", func() error {
			_, err := c.SnapToRoad(ctx, &barikoi.SnapToRoadRequest{Point: "junk"})
			return err
		}},
	}
	for _, tc := range cases {
		err := tc.call()
		var ve *barikoi.ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("%s: error = %T (%v), want *ValidationError", tc.name, err, err)
		}
	}
}

var _ net.Error = timeoutErr{}

var _ = strconv.Itoa // keep strconv if unused after edits
