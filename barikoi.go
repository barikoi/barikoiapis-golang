// Package barikoi is the public entry point of the Barikoi Go SDK for the
// Barikoi Location APIs (https://barikoi.xyz), covering geocoding, routing,
// and place search. It is the Go counterpart of the TypeScript SDK
// barikoiapis (https://www.npmjs.com/package/barikoiapis) and mirrors its
// method names, defaults, validation rules, and error types.
//
// Create a client with an API key obtained from https://developer.barikoi.com:
//
//	c, err := barikoi.NewClient("YOUR_API_KEY")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := c.ReverseGeocode(context.Background(), &barikoi.ReverseGeocodeRequest{
//		Latitude:  23.806703092211507,
//		Longitude: 90.35722628659195,
//		Area:      true,
//	})
//
// The API key is sent as the api_key query parameter on every request, GET
// and POST alike. Use [Client.SetAPIKey] to rotate the key at runtime.
//
// Every method validates its inputs before making an HTTP call and returns
// one of three error types: [*ValidationError] for invalid input,
// [*BarikoiError] for a non-2xx API response, and [*TimeoutError] when the
// request is cancelled or times out. Distinguish them with errors.As.
//
// This package re-exports the SDK surface from the client package, which
// wraps the API client generated from the OpenAPI specification in the
// openapi/ directory. See CONTRIBUTING.md for the architecture.
package barikoi

import (
	"net/http"
	"time"

	"github.com/barikoi/barikoiapis-golang/client"
)

// Defaults used by NewClient, mirroring the TypeScript SDK's BarikoiConfig.
const (
	DefaultBaseURL = client.DefaultBaseURL // "https://barikoi.xyz"
	DefaultTimeout = client.DefaultTimeout // 30s
)

// Client talks to the Barikoi Location APIs. Create one with NewClient; it is
// safe for concurrent use. All API methods hang off *Client.
type Client = client.Client

// Option configures a Client in NewClient.
type Option = client.Option

// NewClient returns a Client that authenticates with the given API key,
// obtained from https://developer.barikoi.com.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	return client.NewClient(apiKey, opts...)
}

// WithBaseURL sets the API base URL (default "https://barikoi.xyz"). The URL
// must use https; pass WithAllowInsecure to permit http for local
// development.
func WithBaseURL(rawURL string) Option { return client.WithBaseURL(rawURL) }

// WithTimeout sets the per-request timeout (default 30s).
func WithTimeout(d time.Duration) Option { return client.WithTimeout(d) }

// WithHTTPClient sets a custom *http.Client for outgoing requests.
func WithHTTPClient(hc *http.Client) Option { return client.WithHTTPClient(hc) }

// WithAllowInsecure permits http:// base URLs, refused by default. Use only
// for local development and testing.
func WithAllowInsecure() Option { return client.WithAllowInsecure() }

// Errors, distinguished with errors.As.
type (
	// BarikoiError is returned for any non-2xx HTTP response from the API.
	BarikoiError = client.BarikoiError
	// ValidationError is returned when a request fails client-side
	// validation, before any HTTP call is made.
	ValidationError = client.ValidationError
	// TimeoutError is returned when a request is cancelled or times out.
	TimeoutError = client.TimeoutError
)

// Sentinel errors returned by the client.
var (
	// ErrMissingAPIKey is returned by NewClient when the API key is empty.
	ErrMissingAPIKey = client.ErrMissingAPIKey
	// ErrInvalidLatitude is returned when a latitude is outside [-90, 90].
	ErrInvalidLatitude = client.ErrInvalidLatitude
	// ErrInvalidLongitude is returned when a longitude is outside [-180, 180].
	ErrInvalidLongitude = client.ErrInvalidLongitude
)

// Shared value types.
type (
	Coordinate        = client.Coordinate
	FlexFloat         = client.FlexFloat
	FlexString        = client.FlexString
	GeoJSONLineString = client.GeoJSONLineString
	RouteGeometry     = client.RouteGeometry
)

// Geocoding methods.
type (
	ReverseGeocodeRequest  = client.ReverseGeocodeRequest
	ReverseGeocodePlace    = client.ReverseGeocodePlace
	ReverseGeocodeResponse = client.ReverseGeocodeResponse
	AutocompleteRequest    = client.AutocompleteRequest
	AutocompletePlace      = client.AutocompletePlace
	AutocompleteResponse   = client.AutocompleteResponse
	GeocodeRequest         = client.GeocodeRequest
	GeocodedPlace          = client.GeocodedPlace
	GeocodeResponse        = client.GeocodeResponse
)

// Search methods.
type (
	SearchPlaceRequest   = client.SearchPlaceRequest
	SearchPlaceResult    = client.SearchPlaceResult
	SearchPlaceResponse  = client.SearchPlaceResponse
	PlaceDetailsRequest  = client.PlaceDetailsRequest
	PlaceDetailsPlace    = client.PlaceDetailsPlace
	PlaceDetailsResponse = client.PlaceDetailsResponse
	NearbyRequest        = client.NearbyRequest
	NearbyPlace          = client.NearbyPlace
	NearbyResponse       = client.NearbyResponse
	CheckNearbyRequest   = client.CheckNearbyRequest
	CheckNearbyPlace     = client.CheckNearbyPlace
	CheckNearbyResponse  = client.CheckNearbyResponse
)

// Routing methods.
type (
	RouteOverviewRequest  = client.RouteOverviewRequest
	RouteLeg              = client.RouteLeg
	Route                 = client.Route
	Waypoint              = client.Waypoint
	RouteOverviewResponse = client.RouteOverviewResponse
	CalculateRouteRequest = client.CalculateRouteRequest
	RouteInstruction      = client.RouteInstruction
	RoutePath             = client.RoutePath
	RoutingResponse       = client.RoutingResponse
	OptimizeRoutePoint    = client.OptimizeRoutePoint
	OptimizeRouteRequest  = client.OptimizeRouteRequest
	SnapToRoadRequest     = client.SnapToRoadRequest
	SnapToRoadResponse    = client.SnapToRoadResponse
)
