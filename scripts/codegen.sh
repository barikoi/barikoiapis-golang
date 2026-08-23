#!/usr/bin/env bash
# Regenerates gen/ from the OpenAPI spec — the Go equivalent of the TypeScript
# repo's `npm run codegen`. Source of truth: openapi/barikoi-api-spec.yaml.
#
# Usage: make codegen   (or scripts/codegen.sh)
#
# The generator is pinned to a specific version so output is deterministic.
# It runs in Docker by default so contributors don't need a local install;
# set OAPI_CODEGEN to a local binary path to use one instead.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION=v2.4.1
IMAGE=golang:1.22

if [[ -n "${OAPI_CODEGEN:-}" ]]; then
    "$OAPI_CODEGEN" -config openapi/cfg.yaml openapi/barikoi-api-spec.yaml
else
    docker run --rm -v "$PWD":/app -w /app "$IMAGE" sh -c \
        "go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$VERSION && \$GOPATH/bin/oapi-codegen -config openapi/cfg.yaml openapi/barikoi-api-spec.yaml"
fi

# Post-generation fixes (mirrors scripts/fix-generated.js in barikoiapis):
#
# 1. oapi-codegen maps OpenAPI `number` to float32, but the TypeScript SDK
#    sends full float64 precision for coordinates. Rewrite float32 -> float64
#    so both SDKs produce identical wire requests.
# 2. geocode_success declares both "Address"/"address" and "subType"/
#    "sub_type" properties, which cannot map to distinct Go fields. Drop the
#    camelCase duplicates (the SDK decodes Rupantor responses itself).
fixup() {
    if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' -e 's/float32/float64/g' \
            -e '/json:"Address,omitempty"/d' \
            -e '/json:"subType,omitempty"/d' gen/gen.go
    else
        sed -i -e 's/float32/float64/g' \
            -e '/json:"Address,omitempty"/d' \
            -e '/json:"subType,omitempty"/d' gen/gen.go
    fi
}
fixup
if ! command -v gofmt >/dev/null 2>&1; then
    docker run --rm -v "$PWD":/app -w /app "$IMAGE" gofmt -w gen/gen.go
else
    gofmt -w gen/gen.go
fi

echo "gen/ regenerated. Run 'make check' to verify."
