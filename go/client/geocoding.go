package client

import (
	"context"
	"net/url"
)

// ReverseGeocodeRequest converts coordinates to a human-readable address.
// Each boolean flag opts in to extra fields in the response; every enabled
// flag may consume additional API credits, so request only what you need.
type ReverseGeocodeRequest struct {
	Latitude     float64
	Longitude    float64
	CountryCode  string // two-letter ISO Alpha-2 code, e.g. "bd"
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
	LocationType         string     `json:"location_type"`
	Thana                string     `json:"thana"`
	ThanaBn              string     `json:"thana_bn"`
	AddressComponents    struct {
		House string `json:"house"`
		Road  string `json:"road"`
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

// ReverseGeocode converts coordinates (longitude, latitude) to a
// human-readable address via GET /v2/api/search/reverse/geocode.
func (c *Client) ReverseGeocode(ctx context.Context, req *ReverseGeocodeRequest) (*ReverseGeocodeResponse, error) {
	if err := validateCoords(req.Latitude, req.Longitude); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("latitude", formatFloat(req.Latitude))
	q.Set("longitude", formatFloat(req.Longitude))
	if req.CountryCode != "" {
		q.Set("country_code", req.CountryCode)
	}
	setBoolParam(q, "country", req.Country)
	setBoolParam(q, "district", req.District)
	setBoolParam(q, "post_code", req.PostCode)
	setBoolParam(q, "sub_district", req.SubDistrict)
	setBoolParam(q, "union", req.Union)
	setBoolParam(q, "pauroshova", req.Pauroshova)
	setBoolParam(q, "location_type", req.LocationType)
	setBoolParam(q, "division", req.Division)
	setBoolParam(q, "address", req.Address)
	setBoolParam(q, "area", req.Area)
	setBoolParam(q, "bangla", req.Bangla)
	setBoolParam(q, "thana", req.Thana)

	var resp ReverseGeocodeResponse
	if err := c.doGet(ctx, "/v2/api/search/reverse/geocode", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AutocompleteRequest describes a partial place query. Q is required.
type AutocompleteRequest struct {
	Q      string
	Bangla bool
	City   string // restrict suggestions to a city, e.g. "dhaka"
}

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
	PostCode  FlexString `json:"postCode"`
	PType     string     `json:"pType"`
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
	q := url.Values{}
	q.Set("q", req.Q)
	setBoolParam(q, "bangla", req.Bangla)
	if req.City != "" {
		q.Set("city", req.City)
	}

	var resp AutocompleteResponse
	if err := c.doGet(ctx, "/v2/api/search/autocomplete/place", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GeocodeRequest formats and geocodes an address string (the Rupantor
// geocoder). Q is required; the boolean flags request extra fields.
type GeocodeRequest struct {
	Q        string
	Thana    bool
	District bool
	Bangla   bool
}

// GeocodedPlace is the matched place returned by Geocode.
type GeocodedPlace struct {
	ID                int64      `json:"id"`
	UCode             string     `json:"uCode"`
	PlaceCode         string     `json:"place_code"`
	Address           string     `json:"address"`
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
	Longitude         FlexFloat  `json:"longitude"`
	Latitude          FlexFloat  `json:"latitude"`
	GeoLocation       []float64  `json:"geo_location"` // [longitude, latitude]
	PopularityRanking int        `json:"popularity_ranking"`
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
// accepts form-encoded bodies, so the request is sent as such even though the
// rest of the SDK uses JSON.
func (c *Client) Geocode(ctx context.Context, req *GeocodeRequest) (*GeocodeResponse, error) {
	if err := requireString("q", req.Q); err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("q", req.Q)
	setYesFormValue(form, "thana", req.Thana)
	setYesFormValue(form, "district", req.District)
	setYesFormValue(form, "bangla", req.Bangla)

	var resp GeocodeResponse
	if err := c.doPostForm(ctx, "/v2/api/search/rupantor/geocode", nil, form, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// setYesFormValue sets name to "yes" only when v is true, as the Rupantor
// geocoder expects.
func setYesFormValue(form url.Values, name string, v bool) {
	if v {
		form.Set(name, "yes")
	}
}
