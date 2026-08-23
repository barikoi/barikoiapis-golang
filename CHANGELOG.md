# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-08-23

### Added

- Initial public API: `NewClient` with `WithBaseURL`, `WithTimeout`,
  `WithHTTPClient`, and `WithAllowInsecure` options; `SetAPIKey`/`GetAPIKey`
  and `SetTimeout`/`GetTimeout` runtime management.
- All 11 Barikoi endpoints, mirroring the TypeScript SDK's method names,
  defaults, and validation rules:
  `ReverseGeocode`, `Autocomplete`, `Geocode` (Rupantor), `SearchPlace`,
  `PlaceDetails`, `Nearby`, `CheckNearby`, `RouteOverview`, `CalculateRoute`,
  `OptimizeRoute`, `SnapToRoad`.
- HTTP client generated from the OpenAPI specification (`gen/`,
  `make codegen`) with a hand-written wrapper (`client/`) and public facade
  (`barikoi.go`).
- Error types matching the TypeScript SDK: `BarikoiError`
  (`IsAuthError`/`IsRateLimitError`/`IsServerError`), `ValidationError`,
  `TimeoutError`, plus `ErrMissingAPIKey`, `ErrInvalidLatitude`,
  `ErrInvalidLongitude`.
- Tolerant response decoding (`FlexFloat`, `FlexString`, `RouteGeometry`)
  for the live API's number/string and polyline/GeoJSON inconsistencies.
- Unit tests against httptest stubs and live integration tests that skip
  without `BARIKOI_API_KEY` (`make test`, `make test-integration`).
