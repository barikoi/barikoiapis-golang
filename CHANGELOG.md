# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Every change documents itself under a proper `major.minor.patch` heading
when it lands; `scripts/release.sh` releases whatever version the
changelog top section names.

## [1.0.6] - 2026-08-23

### Changed

- `Route.Geometry` (RouteOverview) is now the shared `RouteGeometry` type
  (`Polyline` string / `GeoJSON *GeoJSONLineString`) instead of the raw
  `PolylineOrGeoJSON` string blob, matching how `CalculateRoute` and
  `OptimizeRoute` already expose geometries. `PolylineOrGeoJSON` has been
  removed — one geometry type for all routing endpoints.
- `AutocompleteRequest.Bangla` is now a plain `bool` instead of `*bool`,
  matching every other request struct in the SDK. `Bangla: true` now works
  directly — no `barikoi.BoolPtr` needed. The `BoolPtr` helper has been
  removed from the public API. The flag is sent as-is on every request, so
  users wanting Bangla fields must set `Bangla: true` explicitly (the
  TypeScript SDK defaults it to `true`, which a Go zero-value `bool` cannot
  express).

## [1.0.5] - 2026-08-23

### Changed

- CI: CodeQL moved to its own workflow (`codeql.yml`), enabling a dedicated
  README badge (per-job badges don't exist on GitHub Actions).
- README: added the CodeQL badge.

## [1.0.4] - 2026-08-23

### Added

- Runnable godoc examples (`example_test.go`) for the package overview,
  `Client.Geocode`, and `Client.Autocomplete`, surfaced on pkg.go.dev.

### Changed

- Release script now warms the Go module proxy after tagging, so pkg.go.dev
  picks up the new tagged/stable version immediately.

## [1.0.3] - 2026-08-23

### Changed

- CI: dropped the redundant `github.event_name == 'push'` check from the
  coverage-badge publish condition; the `refs/heads/main` ref check alone
  is conclusive given the workflow's triggers.

## [1.0.2] - 2026-08-23

### Changed

- Release process: changes are now documented under a proper version
  heading in `CHANGELOG.md` when they land (no `[Unreleased]` staging
  section). `scripts/release.sh` reads the top version heading from the
  changelog, updates `dev`, merges `dev` into `main` with an explicit
  merge commit, tags, and opens the GitHub release from the changelog
  notes.

## [1.0.1] - 2026-08-23

### Added

- CI workflow (`.github/workflows/ci.yml`): `make lint`, unit tests with a
  coverage profile, CodeQL analysis, and golangci-lint (`gen/` excluded via
  `.golangci.yml`).
- Self-hosted coverage badge: CI publishes a shields-endpoint JSON to the
  `badges` branch on every push to main — no third-party coverage service
  or token needed.
- `scripts/release.sh`: release helper enforcing strict `major.minor.patch`
  versioning, folding the changelog `[Unreleased]` section into the new
  version, tagging, pushing `dev` -> `main`, and opening the GitHub
  release from the changelog notes.

### Fixed

- Check ignored error returns flagged by golangci-lint errcheck
  (`resp.Body.Close` in `client/client.go`; `.env` loading in
  `test/integration_test.go`).

### Changed

- README rewritten: per-method type declarations, descriptions mirroring
  the TypeScript SDK, Go-native `net/http` transport notes, environment-
  based API key examples, and flat-square badges (Go Reference, Release,
  CI, Coverage, Go Report Card, Go Version, License).

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
