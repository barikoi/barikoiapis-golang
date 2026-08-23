package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/barikoi/barikoiapis-golang/gen"
)

// ReverseGeocodeRequest converts coordinates to a human-readable address.
// Each boolean flag opts in to extra fields in the response; every enabled
// flag may consume additional API credits, so request only what you need.
// CountryCode defaults to "BD" when empty.
type ReverseGeocodeRequest struct {
	Latitude     float64
	Longitude    float64
	CountryCode  string // two-letter ISO Alpha-2 code, e.g. "BD"
	Country      bool
	District     bool
	PostCode     bool
	SubDistrict  bool
	Union        bool
	Pauroshova   bool
	LocationType bool
	Division     bool
	Address      bool
	Area         bool
	Bangla       bool
	Thana        bool
}

// ReverseGeocodePlace is the place returned by ReverseGeocode.
type ReverseGeocodePlace struct {
	ID                   int64      `json:"id"`
	DistanceWithinMeters FlexFloat  `json:"distance_within_meters"`
	Address              string     `json:"address"`
	Area                 string     `json:"area"`
	City                 string     `json:"city"`
	PostCode             FlexString `json:"postCode"`
	AddressBn            string     `json:"address_bn"`
	AreaBn               string     `json:"area_bn"`
	CityBn               string     `json:"city_bn"`
	Country              string     `json:"country"`
	Division             string     `json:"division"`
	District             string     `json:"district"`
	SubDistrict          string     `json:"sub_district"`
	Union                string     `json:"union"`
	Pauroshova           string     `json:"pauroshova"`
	LocationType         string     `json:"location_type"`
	Thana                string     `json:"thana"`
	ThanaBn              string     `json:"thana_bn"`
	AddressComponents    struct {
		PlaceName string `json:"place_name"`
		House     string `json:"house"`
		Road      string `json:"road"`
	} `json:"address_components"`
	AreaComponents struct {
		Area    string `json:"area"`
		SubArea string `json:"sub_area"`
	} `json:"area_components"`
}

// ReverseGeocodeResponse is the response of ReverseGeocode.
type ReverseGeocodeResponse struct {
	Place  ReverseGeocodePlace `json:"place"`
	Status int                 `json:"status"`
}

// ReverseGeocode converts coordinates (latitude, longitude) to a
// human-readable address via GET /v2/api/search/reverse/geocode.
func (c *Client) ReverseGeocode(ctx context.Context, req *ReverseGeocodeRequest) (*ReverseGeocodeResponse, error) {
	if err := validateCoords(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}
	countryCode := req.CountryCode
	if countryCode == "" {
		countryCode = "BD" // default, matching the TypeScript SDK
	}
	params := &gen.ReverseGeocodeParams{
		ApiKey:       c.apiKeyParam(),
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		CountryCode:  &countryCode,
		Country:      optBool(req.Country),
		District:     optBool(req.District),
		PostCode:     optBool(req.PostCode),
		SubDistrict:  optBool(req.SubDistrict),
		Union:        optBool(req.Union),
		Pauroshova:   optBool(req.Pauroshova),
		LocationType: optBool(req.LocationType),
		Division:     optBool(req.Division),
		Address:      optBool(req.Address),
		Area:         optBool(req.Area),
		Bangla:       optBool(req.Bangla),
		Thana:        optBool(req.Thana),
	}

	var resp ReverseGeocodeResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) { return c.gen.ReverseGeocode(ctx, params) }, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AutocompleteRequest describes a partial place query. Q is required.
// Bangla defaults to true when nil, matching the TypeScript SDK; set it to
// false explicitly (BoolPtr(false)) to omit Bangla fields.
type AutocompleteRequest struct {
	Q      string
	Bangla *bool
}

// BoolPtr returns a pointer to v, for optional boolean fields such as
// AutocompleteRequest.Bangla.
func BoolPtr(v bool) *bool { return &v }

// AutocompletePlace is one suggestion returned by Autocomplete.
type AutocompletePlace struct {
	ID        int64      `json:"id"`
	Longitude FlexFloat  `json:"longitude"`
	Latitude  FlexFloat  `json:"latitude"`
	Address   string     `json:"address"`
	AddressBn string     `json:"address_bn"`
	City      string     `json:"city"`
	CityBn    string     `json:"city_bn"`
	Area      string     `json:"area"`
	AreaBn    string     `json:"area_bn"`
	District  string     `json:"district"`
	PostCode  FlexString `json:"postCode"`
	PType     string     `json:"pType"`
	SubType   string     `json:"subType"`
	UCode     string     `json:"uCode"`
}

// AutocompleteResponse is the response of Autocomplete.
type AutocompleteResponse struct {
	Places []AutocompletePlace `json:"places"`
	Status int                 `json:"status"`
}

// Autocomplete returns place suggestions for a partial query via
// GET /v2/api/search/autocomplete/place.
func (c *Client) Autocomplete(ctx context.Context, req *AutocompleteRequest) (*AutocompleteResponse, error) {
	if err := requireString("q", req.Q); err != nil {
		return nil, err
	}
	bangla := req.Bangla == nil || *req.Bangla
	params := &gen.AutocompleteParams{
		ApiKey: c.apiKeyParam(),
		Q:      req.Q,
		Bangla: &bangla,
	}

	var resp AutocompleteResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) { return c.gen.Autocomplete(ctx, params) }, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GeocodeRequest formats and geocodes an address string (the Rupantor
// geocoder). Q is required; the boolean flags request extra fields and are
// sent as "yes" only when true, matching the TypeScript SDK's omission
// behavior.
type GeocodeRequest struct {
	Q        string
	Thana    bool
	District bool
	Bangla   bool
}

// GeocodedPlace is the matched place returned by Geocode. The API returns
// the address under an "address" or "Address" key depending on the query;
// both are accepted and Address is populated from either.
type GeocodedPlace struct {
	ID                int64      `json:"id"`
	UCode             string     `json:"uCode"`
	PlaceCode         string     `json:"place_code"`
	Address           string     `json:"address"`
	AddressTitle      string     `json:"Address"`
	AddressBn         string     `json:"address_bn"`
	BusinessName      string     `json:"business_name"`
	Area              string     `json:"area"`
	AreaBn            string     `json:"area_bn"`
	SubArea           string     `json:"sub_area"`
	SuperSubArea      string     `json:"super_sub_area"`
	City              string     `json:"city"`
	CityBn            string     `json:"city_bn"`
	District          string     `json:"district"`
	SubDistrict       string     `json:"sub_district"`
	Thana             string     `json:"thana"`
	PType             string     `json:"pType"`
	SubType           string     `json:"subType"`
	PostCode          FlexString `json:"postCode"`
	Postcode          FlexString `json:"postcode"`
	Longitude         FlexFloat  `json:"longitude"`
	Latitude          FlexFloat  `json:"latitude"`
	GeoLocation       []float64  `json:"geo_location"` // [longitude, latitude]
	PopularityRanking int        `json:"popularity_ranking"`
}

// UnmarshalJSON fills Address from the "address" key, falling back to the
// "Address" key that some Rupantor responses use instead.
func (p *GeocodedPlace) UnmarshalJSON(data []byte) error {
	type geocodedPlaceAlias GeocodedPlace
	var a geocodedPlaceAlias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*p = GeocodedPlace(a)
	if p.Address == "" {
		p.Address = p.AddressTitle
	}
	return nil
}

// GeocodeResponse is the response of Geocode.
type GeocodeResponse struct {
	GivenAddress              string        `json:"given_address"`
	FixedAddress              string        `json:"fixed_address"`
	BanglaAddress             string        `json:"bangla_address"`
	AddressStatus             string        `json:"address_status"`
	GeocodedAddress           GeocodedPlace `json:"geocoded_address"`
	ConfidenceScorePercentage FlexFloat     `json:"confidence_score_percentage"`
	Status                    int           `json:"status"`
}

// Geocode formats and geocodes an address string to coordinates via the
// Rupantor Geocoder, POST /v2/api/search/rupantor/geocode. The endpoint only
// accepts form-encoded bodies. One Rupantor request consumes two Geocode API
// credits.
func (c *Client) Geocode(ctx context.Context, req *GeocodeRequest) (*GeocodeResponse, error) {
	if err := requireString("q", req.Q); err != nil {
		return nil, err
	}
	params := &gen.GeocodeParams{ApiKey: c.apiKeyParam()}
	body := gen.GeocodeBody{
		Q:        req.Q,
		Bangla:   yesNo[gen.GeocodeBodyBangla](req.Bangla),
		District: yesNo[gen.GeocodeBodyDistrict](req.District),
		Thana:    yesNo[gen.GeocodeBodyThana](req.Thana),
	}

	var resp GeocodeResponse
	err := c.do(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.gen.GeocodeWithFormdataBody(ctx, params, body)
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// optBool returns a pointer to v for optional boolean query parameters, or
// nil when v is false so the parameter is omitted — matching the TypeScript
// SDK, which omits boolean flags the caller did not set.
func optBool(v bool) *bool {
	if !v {
		return nil
	}
	return &v
}

// yesNo maps a boolean flag to the "yes" enum value the Rupantor geocoder
// expects, or nil to omit the field when the flag is false.
func yesNo[T ~string](v bool) *T {
	if !v {
		return nil
	}
	yes := T("yes")
	return &yes
}
