# Contributing to barikoiapis-golang

Single source of truth for working on this library: setup, architecture,
code generation, testing, gotchas, and release checklist.

For user-facing API docs see [README.md](./README.md). Agent-specific
instructions live in [AGENTS.md](./AGENTS.md).

This SDK is the Go counterpart of the TypeScript SDK
[barikoiapis](https://www.npmjs.com/package/barikoiapis); architecture and
behavior mirror it where Go allows.

---

## Quick Start

1. **Setup:** see [Development Environment](#development-environment) below.
2. **Workflow:**
   - Branch from `main` for each change.
   - Write tests for new features and bugfixes.
   - Run `make check` before opening a PR.
3. **Merges:** maintainers merge feature branches into `main`.

---

## Development Environment

### Prerequisites

- Go >= 1.22 **or** Docker (the Makefile transparently builds/tests in
  `golang:1.22` when no local toolchain exists)
- GNU Make
- Git

### Setup

```bash
git clone https://github.com/barikoi/barikoiapis-golang.git
cd barikoiapis-golang
make test
```

The generated client (`gen/`) is committed, so no codegen step is needed to
build and test. Only run `make codegen` when the OpenAPI spec changes.

### Environment variables

Live API access needs a key from
[developer.barikoi.com](https://developer.barikoi.com):

```bash
cp env.example .env        # then edit .env with your key
```

`.env` is gitignored; integration tests and the example read
`BARIKOI_API_KEY` from the environment or `.env` automatically. **Never
commit a key.**

### Available commands

| Command | Description |
|--------|-------------|
| `make build` | Compile all packages |
| `make test` | Unit tests with race detector and coverage (no key needed) |
| `make test-integration` | Live API tests; requires `BARIKOI_API_KEY` |
| `make cover` | Unit tests + coverage summary |
| `make lint` | gofmt check + `go vet` |
| `make check` | `lint` + `test` — run before every PR |
| `make codegen` | Regenerate `gen/` from `openapi/barikoi-api-spec.yaml` |
| `make example` | Run `examples/basic` against the live API |
| `make tidy` | `go mod tidy` |

### Project Structure

```
openapi/
  barikoi-api-spec.yaml  # Source of truth — copy of the barikoiapis spec
  cfg.yaml               # oapi-codegen config (≈ openapi-ts.config.ts)
gen/
  gen.go                 # AUTO-GENERATED from the spec — never hand-edit
client/                  # Hand-written SDK (≈ src/lib/ in barikoiapis)
  client.go              # Client, options, timeout, error mapping
  errors.go              # BarikoiError, ValidationError, TimeoutError
  validation.go          # Coordinate/format/enum validation (≈ Zod schemas)
  geocoding.go           # ReverseGeocode, Autocomplete, Geocode
  search.go              # SearchPlace, PlaceDetails, Nearby, CheckNearby
  routing.go             # RouteOverview, CalculateRoute, OptimizeRoute, SnapToRoad
  *_test.go              # Unit tests (httptest stubs)
  integration_test.go    # Live API tests; auto-skip without BARIKOI_API_KEY
barikoi.go               # Public facade — re-exports client/ (≈ src/index.ts)
examples/basic/          # Runnable program exercising all 11 endpoints
scripts/
  codegen.sh             # Regeneration + post-generation fixes
Makefile                 # Developer commands (≈ package.json scripts)
```

---

## Architecture

### Generated vs hand-written code

Mirrors the TypeScript SDK's split:

| barikoiapis (TS) | this repo (Go) | Role |
|---|---|---|
| `openapi/barikoi-api-spec.yaml` | `openapi/barikoi-api-spec.yaml` | Source of truth |
| `openapi-ts.config.ts` | `openapi/cfg.yaml` | Generator config |
| `@hey-api/openapi-ts` | `oapi-codegen` (pinned v2.4.1) | Generator |
| `src/client/*.gen.ts` | `gen/gen.go` | Generated models + HTTP methods |
| `src/lib/client.ts` | `client/client.go` | Wrapper: validation, API key, timeout |
| `src/lib/validation.ts` | `client/validation.go` | Input validation |
| `src/lib/errors.ts` | `client/errors.go` | Error types |
| `src/index.ts` | `barikoi.go` | Public exports |
| `npm run codegen` | `make codegen` | Regenerate |
| `npm test` | `make test` | Tests |

- `gen/` is auto-generated. **Never edit it by hand.** To change an endpoint
  signature, edit the OpenAPI spec and run `make codegen`.
- `client/` validates inputs (before any HTTP call), applies TypeScript-SDK
  defaults, injects the API key, applies timeouts, and decodes responses
  into tolerant public types.
- `barikoi.go` re-exports the public surface with type aliases, so consumers
  import only `github.com/barikoi/barikoiapis-golang`.

### Why responses decode in `client/`, not `gen/`

The live API is looser than the spec: coordinates and post codes arrive as
JSON numbers *or* strings depending on the endpoint; Rupantor returns the
address under `"address"` or `"Address"`; nearby uses snake_case keys absent
from the spec. `FlexFloat`/`FlexString` and custom `UnmarshalJSON` methods
absorb this, so every response field is a plain Go value.

### Known API drift (spec vs live) — intentional, do not "fix"

| Endpoint | Spec says | Live API does | SDK behavior |
|---|---|---|---|
| CalculateRoute response | GraphHopper `{hints, info, paths}` | Same | GraphHopper shape (the old Valhalla `trip` model was wrong) |
| Nearby place fields | `pType`/`subType`/`postCode` camelCase | `type`/`sub_type`/`postcode` snake_case, `id` numeric string | Live shape |
| Rupantor geocoded address | `address` | `address` or `Address` per query | Both accepted |
| optimizeRoute auth | `api_key` in body only | Query param expected (TS injects it) | Sent as both |

---

## Code Generation

### When to regenerate

Run `make codegen` whenever `openapi/barikoi-api-spec.yaml` changes — new
endpoint, changed parameter, changed response shape. The script:

1. Runs `oapi-codegen` (pinned v2.4.1, Docker-based by default; set
   `OAPI_CODEGEN=/path/to/oapi-codegen` to use a local binary).
2. Post-generation fixes (≈ `scripts/fix-generated.js` in barikoiapis):
   - `float32` → `float64`: oapi maps OpenAPI `number` to float32, which
     loses coordinate precision the TypeScript SDK keeps on the wire.
   - Drops duplicate Go fields from `GeocodeSuccess` (`Address`/`address`,
     `subType`/`sub_type` collide in Go; the SDK decodes Rupantor itself).
3. `gofmt`s the result.

### Post-generation checklist

- `make check` — generated code compiles, unit tests still pass.
- Adapt `client/` method wiring if generated types changed.
- `make test-integration` — confirm live responses still decode.

---

## Testing Workflow

### Unit tests

- **Framework:** the standard `testing` package + `net/http/httptest`.
- **Location:** `client/*_test.go`, mirroring the source layout.
- **Network:** none. Stubs assert the exact wire format (query params, path,
  body, content type), catching codegen and encoding regressions.

```bash
make test                                  # full suite
docker run --rm -v "$PWD":/app -w /app golang:1.22 go test -run TestNearby ./client/   # single test without local Go
```

### Integration tests

`client/integration_test.go` hits the **live API**. Tests are named
`TestIntegration*` and skip automatically when `BARIKOI_API_KEY` is unset
(≈ the TypeScript `integrationSuite` helper):

```bash
cp env.example .env   # set your key, then:
make test-integration
```

### Adding an endpoint

1. Update the spec in the **barikoiapis repo**, copy it here, `make codegen`.
2. Wire the method in `client/`: request struct → validation → gen params →
   `c.do` with a decode type.
3. Add a stub unit test and a `TestIntegration*` test.
4. Re-export types in `barikoi.go`.
5. Add a README section.

---

## Gotchas

- **`gen/` is generated.** Never hand-edit; edit the spec and regenerate.
- **oapi-codegen version is pinned (v2.4.1)** for deterministic output and
  Go 1.22 compatibility (v2.8+ requires Go >= 1.25).
- **Unit tests never need a key or network**; keep it that way.
- **httptest speaks plain http**, so tests construct clients with
  `WithAllowInsecure()`; production base URLs must be https.
- **Two runtime dependencies** (`oapi-codegen/runtime`, `google/uuid`) come
  from the generated client; everything else is standard library.

---

## Commit Guidelines

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`.
Scopes are optional (e.g. `feat(routing)`, `fix(validation)`).

```
feat(routing): add route optimization method
fix(validation): correct nearby radius bounds
docs: expand code generation section
```

---

## Release Process

1. Verify with a real key:
   ```bash
   make check && make test-integration
   ```
2. Add a `## [x.y.z] - YYYY-MM-DD` entry to `CHANGELOG.md`.
3. Tag and push:
   ```bash
   git tag vx.y.z
   git push origin main --tags
   ```
4. Consumers pick the release via the Go module proxy; verify with
   `go get github.com/barikoi/barikoiapis-golang@vx.y.z` in a fresh project.
5. Create a GitHub release with the CHANGELOG entry.

Breaking changes bump the major version, which requires a `/vN` module path
change (`barikoiapis-golang/v2`) — avoid unless necessary.

---

## Resources

- [README.md](./README.md) — user-facing API docs
- [AGENTS.md](./AGENTS.md) — instructions for AI coding agents
- [CHANGELOG.md](./CHANGELOG.md) — version history
- [TypeScript SDK](https://www.npmjs.com/package/barikoiapis) — the SDK this mirrors
- [Barikoi API docs](https://docs.barikoi.com/docs/maps-api)
- [oapi-codegen docs](https://github.com/oapi-codegen/oapi-codegen)
- [Conventional Commits](https://www.conventionalcommits.org/)
