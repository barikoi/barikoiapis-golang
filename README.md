# Barikoi Go SDK

Native Go client for the [Barikoi Location APIs](https://barikoi.xyz) — geocoding, routing, and place search for Bangladesh. Standard library only, no third-party dependencies.

## Installation

```bash
go get github.com/barikoi/barikoiapis-golang
```

```go
import bk "github.com/barikoi/barikoiapis-golang/client"
```

Get an API key at [developer.barikoi.com](https://developer.barikoi.com/).

## Quick start

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	bk "github.com/barikoi/barikoiapis-golang/client"
)

func main() {
	c, err := bk.NewClient("YOUR_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.ReverseGeocode(context.Background(), &bk.ReverseGeocodeRequest{
		Latitude:  23.806703092211507,
		Longitude: 90.35722628659195,
		Area:      true,
	})
	if err != nil {
		var apiErr *bk.BarikoiError
		if errors.As(err, &apiErr) && apiErr.IsAuthError() {
			log.Fatal("invalid API key")
		}
		log.Fatal(err)
	}
	fmt.Println(resp.Place.Address)
}
```

The API key is sent as the `api_key` query parameter on every request, GET and POST alike. Rotate it at runtime with `c.SetAPIKey("new-key")` / `c.GetAPIKey()`.

## Methods

| Method | Endpoint | Description |
| --- | --- | --- |
| `ReverseGeocode` | `GET /v2/api/search/reverse/geocode` | Coordinates to human-readable address, with optional fields for district, thana, post code, Bangla output, etc. |
| `Autocomplete` | `GET /v2/api/search/autocomplete/place` | Place suggestions for a partial query as the user types. |
| `Geocode` | `POST /v2/api/search/rupantor/geocode` | Formats and geocodes an address string to coordinates (Rupantor geocoder; form-encoded body). |
| `RouteOverview` | `GET /v2/api/route/{coordinates}` | Basic route (distance, duration, geometry) between `lon,lat;lon,lat` points. |
| `CalculateRoute` | `POST /v2/api/routing` | Detailed route with turn-by-turn maneuvers between two coordinates. |
| `OptimizeRoute` | `POST /v2/api/route/optimized` | Optimized route through up to 50 waypoints (`source`, `destination`, `geo_points`). |
| `SnapToRoad` | `GET /v2/api/routing/nearest` | Snaps a coordinate (`lat,lon`) to the nearest point on the road network. |
| `SearchPlace` | `GET /api/v2/search-place` | Search places by query; returns place codes and a session ID. |
| `PlaceDetails` | `GET /api/v2/places` | Address and coordinates for a place code, using the session ID from `SearchPlace`. |
| `Nearby` | `GET /v2/api/search/nearby/{radius}/{limit}` | Places near a point; radius (km) and limit are path parameters. |
| `CheckNearby` | `GET /v2/api/check/nearby` | Whether the destination point is within a radius (10–1000 m) of the current point. |

All methods take a `context.Context` first and validate their inputs before making any HTTP call.

## Configuration

```go
c, err := bk.NewClient(
	"YOUR_API_KEY",
	bk.WithBaseURL("https://barikoi.xyz"), // default
	bk.WithTimeout(30*time.Second),        // default 30s
	bk.WithHTTPClient(&http.Client{...}),  // default: shared client with the timeout above
)
```

- `WithBaseURL` — API base URL; useful for proxies and tests.
- `WithTimeout` — per-request timeout; overrides a custom client's `Timeout`.
- `WithHTTPClient` — custom `*http.Client`; it keeps its own `Timeout` unless `WithTimeout` is also given.

## Error handling

Three error types, distinguished with `errors.As`:

- `*bk.BarikoiError` — any non-2xx API response. Fields: `Message`, `StatusCode`, `Code`, `Details` (raw body). Predicates: `IsAuthError()` (401/403), `IsRateLimitError()` (429), `IsServerError()` (5xx).
- `*bk.ValidationError` — client-side validation failure (`Field`, `Message`); no HTTP call was made. `bk.ErrInvalidLatitude` and `bk.ErrInvalidLongitude` are predefined instances.
- `*bk.TimeoutError` — the request was cancelled or timed out.

```go
_, err := c.Nearby(ctx, &bk.NearbyRequest{Latitude: 23.8, Longitude: 90.3, Radius: 5, Limit: 10})
var apiErr *bk.BarikoiError
if errors.As(err, &apiErr) && apiErr.IsRateLimitError() {
	// back off and retry
}
var valErr *bk.ValidationError
if errors.As(err, &valErr) {
	fmt.Println("bad field:", valErr.Field) // "radius", "limit", ...
}
```

## Validation rules

- Latitude in `[-90, 90]`, longitude in `[-180, 180]` → `ErrInvalidLatitude` / `ErrInvalidLongitude`.
- Required strings (`Q`, `PlaceCode`, `SessionID`, `Coordinates`, `Point`, `Source`, `Destination`) → `*ValidationError` naming the field.
- `Nearby`: radius in `[0.1, 100]` km, limit in `[1, 100]`.
- `CheckNearby`: radius in `[10, 1000]` meters.

## Tolerant response types

Barikoi returns coordinates and post codes as JSON numbers in some endpoints and as strings in others. The SDK's `FlexFloat` and `FlexString` types unmarshal either form, so every response field is a plain Go `float64`-like or `string`-like value regardless of endpoint.

## Endpoints and caveats

All 11 endpoints above were verified against the official API reference at <https://docs.barikoi.com/api/> (checked August 2026). Caveats worth knowing:

- **`Geocode`** maps to the Rupantor Geocoder (`/v2/api/search/rupantor/geocode`), the only documented forward-geocoding endpoint. It only accepts `application/x-www-form-urlencoded` bodies, so the SDK sends a form body there even though other POSTs use JSON. Its boolean flags go out as `yes`, per the API. One Rupantor request consumes two Geocode API credits, per the docs.
- **`OptimizeRoute`** follows the documented request shape (`source`, `destination`, `profile`, `geo_points`), not a single coordinates string. The docs show `api_key` in the JSON body, so the SDK sends it both in the body and as the `api_key` query parameter (as on every request).
- **`CalculateRoute`**: the docs' curl example uses `key=` in the query string, but the parameter table says `api_key`; the SDK sends `api_key` like every other endpoint. Verify with a live key if a 401 occurs.
- **`RouteOverview`** leg `steps` have no documented schema; the SDK exposes them as raw `map[string]any`.
- **`SnapToRoad`** maps to the nearest-road endpoint (`/v2/api/routing/nearest`, param `point`): it snaps a single coordinate to the closest road point. The live API rejects the undocumented map-matching endpoint (`/v2/api/routing/matching`) with 503.
- Not yet exercised against a live API with a real key — response types are taken from the documented examples. If a field arrives in a different shape, `FlexFloat`/`FlexString` absorb number/string mismatches, but new fields may be missing from the structs. File an issue or add the field.

## Example

See [`examples/basic`](examples/basic/main.go) for a runnable program (`BARIKOI_API_KEY=... go run ./examples/basic`).

## Development

```bash
cd go
gofmt -l .
go vet ./...
go test -race -cover ./...
```
