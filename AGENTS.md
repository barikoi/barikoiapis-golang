# AGENTS.md

Instructions for AI coding agents working in this repository. Read this before making any change.

## What this repo is

The official Go SDK for the Barikoi Location APIs — the Go counterpart of
the TypeScript/JavaScript SDK [`barikoiapis`](https://www.npmjs.com/package/barikoiapis).
Method names, defaults, validation rules, and error types must stay in sync
with the TypeScript SDK; developers switching between the two should feel no
difference beyond the language.

## Golden rules

1. **`gen/` is auto-generated. Never hand-edit.** Change the OpenAPI spec
   (`openapi/barikoi-api-spec.yaml`), run `make codegen`, then adapt
   `client/`.
2. **`openapi/` is the source of truth** for endpoints, parameters, request
   bodies, and enums. If the spec and any wrapper code disagree, the spec
   wins — fix the wrapper.
3. **The public API is the root `barikoi` package only** (`barikoi.go`
   re-exports). `client/` and `gen/` are internal despite living as normal
   packages; do not advertise them in the README or examples.
4. **Never commit API keys.** Live keys are passed via the `BARIKOI_API_KEY`
   environment variable or a gitignored `.env` file (see `env.example`).
5. **Run `make check` before declaring work done** (gofmt + vet + unit
   tests). Unit tests need no key and no network.
6. **Response types in `client/` decode the live API, not the spec**, when
   the two differ. Known live-vs-spec drift is intentional and documented in
   `CONTRIBUTING.md` ("Known API drift"). Do not "fix" those back to the
   spec.

## Commands

| Command | Purpose |
|---|---|
| `make build` | Compile all packages |
| `make test` | Unit tests (httptest stubs; no key needed) |
| `make test-integration` | Live API tests; needs `BARIKOI_API_KEY` |
| `make lint` | gofmt check + `go vet` |
| `make check` | lint + test — run before finishing |
| `make codegen` | Regenerate `gen/` from the OpenAPI spec |
| `make example` | Run `examples/basic` against the live API |

The Makefile transparently uses Docker (`golang:1.22`) when no local Go
toolchain is installed, so these commands work on any machine with Docker.

## Architecture (read before touching transport code)

```
openapi/barikoi-api-spec.yaml   source of truth (copy of the barikoiapis spec)
openapi/cfg.yaml                oapi-codegen config (≈ openapi-ts.config.ts)
gen/gen.go                      GENERATED models + HTTP client (≈ src/client/)
client/                         hand-written SDK (≈ src/lib/ in barikoiapis)
  client.go                     Client, options, timeout, error mapping
  errors.go                     BarikoiError, ValidationError, TimeoutError
  validation.go                 coordinate/format/enum validation (≈ Zod schemas)
  geocoding.go search.go routing.go   the 11 API methods
  *_test.go                     unit tests (httptest)
  integration_test.go           live tests, auto-skip without BARIKOI_API_KEY
barikoi.go                      public facade; re-exports client/ types
```

Request flow: `barikoi.X` → `client.Client.X` (validates, applies defaults,
adds `api_key`, wraps the timeout) → `gen` method (builds the request from
the spec: path, query, body, content type) → response decoded in `client/`
into the tolerant public types (`FlexFloat`/`FlexString` absorb the live
API's number/string inconsistencies).

### Codegen post-processing is intentional

`scripts/codegen.sh` rewrites `float32` → `float64` in `gen/gen.go` (oapi's
`number` mapping loses coordinate precision the TS SDK keeps) and drops two
duplicate Go fields from `GeocodeSuccess` (`Address`/`address`,
`subType`/`sub_type` collide in Go). If you regenerate and see those issues
return, the script's fixups are the fix — don't work around them by hand.

### optimizeRoute's api_key

The spec declares `api_key` for optimizeRoute only in the request body, but
the TS generator injects the global apiKey security scheme as a query
parameter on every operation. `client/routing.go` replicates that with a
request editor (`withAPIKeyQuery`). If the spec ever adds the query
parameter, drop the editor.

## Testing conventions

- Unit tests assert the exact wire format (query params, path, body, content
  type) against `httptest` servers — they will catch codegen or encoding
  regressions.
- Integration tests are named `TestIntegration*` and skip automatically
  without `BARIKOI_API_KEY`; never make unit tests depend on a key.
- Adding an endpoint? Follow an existing file: request struct → validation →
  gen params struct → `c.do` with decode type → unit test with stub →
  integration test → facade alias in `barikoi.go` → README section.

## Gotchas

- `client/` speaks plain http to `httptest` servers; tests construct clients
  with `WithAllowInsecure()` because production base URLs must be https.
- `PlaceDetailsRequest.SessionID` is a string in the public API but parsed as
  UUID (spec `format: uuid`); invalid UUIDs fail validation client-side.
- Rupantor geocode returns the address under `"address"` or `"Address"`
  depending on the query; `GeocodedPlace.UnmarshalJSON` handles both.
- Commit style: Conventional Commits (`feat:`, `fix:`, `docs:`, ...).
