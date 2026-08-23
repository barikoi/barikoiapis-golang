#!/usr/bin/env sh
# Release helper for the Barikoi Go SDK.
#
# Changelog-first: changes are documented under a proper `## [X.Y.Z]`
# heading in CHANGELOG.md when they land. This script releases whatever
# strict major.minor.patch version the changelog's top section names.
#
# Usage: ./scripts/release.sh
#
# Flow:
#   1. working tree must be clean, on branch dev
#   2. CHANGELOG.md top heading must be ## [X.Y.Z] (strict semver core),
#      non-empty, and newer than the latest tag
#   3. `make check` must pass (gofmt + go vet + unit tests)
#   4. update dev, merge dev into main (--no-ff merge commit)
#   5. tag vX.Y.Z on main, push, `gh release create` from changelog notes

set -eu

cd "$(dirname "$0")/.."

[ "$(git branch --show-current)" = "dev" ] || { echo "not on branch dev" >&2; exit 1; }
git diff --quiet && git diff --cached --quiet || { echo "working tree not clean" >&2; exit 1; }

NEXT=$(sed -n 's/^## \[\([^]]*\)\].*/\1/p' CHANGELOG.md | head -1)
case "$NEXT" in
  [0-9]*.[0-9]*.[0-9]*) ;;
  *)
    echo "CHANGELOG.md top section is '## [$NEXT]' — document the change under a proper major.minor.patch heading first" >&2
    exit 1
    ;;
esac

CURRENT=$(git tag -l 'v*' | sed 's/^v//' | sort -V | tail -1)
[ -n "$CURRENT" ] || { echo "no tags found" >&2; exit 1; }
if [ "$(printf '%s\n%s\n' "$CURRENT" "$NEXT" | sort -V | tail -1)" != "$NEXT" ] ||
  [ "$CURRENT" = "$NEXT" ]; then
  echo "changelog top version v$NEXT must be newer than latest tag v$CURRENT" >&2
  exit 1
fi

[ -n "$(awk -v v="$NEXT" '
  index($0, "## [" v "]") == 1 { found = 1; next }
  /^## / { found = 0 }
  found && NF
' CHANGELOG.md)" ] || { echo "CHANGELOG.md section [$NEXT] is empty" >&2; exit 1; }

echo "==> releasing v$NEXT (latest tag v$CURRENT)"
make check

echo "==> update dev"
git push -q origin dev

echo "==> merge dev into main"
git checkout -q main
git pull -q origin main
git merge --no-ff dev -m "release: merge dev into main (v$NEXT)"
git push -q origin main

echo "==> tag and release"
git tag "v$NEXT"
git push -q origin "v$NEXT"

# Warm the Go module proxy so pkg.go.dev shows the new tagged/stable version
# promptly (the proxy only ingests a tag when its zip is first fetched).
curl -sf --max-time 10 "https://proxy.golang.org/github.com/barikoi/barikoiapis-golang/@v/v$NEXT.zip" -o /dev/null \
	|| echo "warning: could not warm proxy.golang.org for v$NEXT" >&2

NOTES=$(awk -v v="$NEXT" '
  index($0, "## [" v "]") == 1 { found = 1; next }
  /^## / { found = 0 }
  found
' CHANGELOG.md)
gh release create "v$NEXT" --title "v$NEXT" --notes "$NOTES"

git checkout -q dev
echo "==> released v$NEXT"
