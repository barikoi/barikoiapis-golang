#!/usr/bin/env sh
# Release helper for the Barikoi Go SDK.
#
# Enforces strict major.minor.patch versioning, keeps CHANGELOG.md as the
# single release log, tags, pushes dev -> main, and opens the GitHub
# release with notes extracted from the changelog.
#
# Usage: ./scripts/release.sh patch|minor|major
#
# Flow:
#   1. working tree must be clean, on branch dev
#   2. `make check` must pass (gofmt + go vet + unit tests)
#   3. the [Unreleased] changelog section becomes the new version section
#   4. tag vX.Y.Z, push dev and dev:main, `gh release create`

set -eu

cd "$(dirname "$0")/.."

BUMP="${1:-}"
case "$BUMP" in
  patch | minor | major) ;;
  *) echo "usage: ./scripts/release.sh patch|minor|major" >&2; exit 1 ;;
esac

[ "$(git branch --show-current)" = "dev" ] || { echo "not on branch dev" >&2; exit 1; }
git diff --quiet && git diff --cached --quiet || { echo "working tree not clean" >&2; exit 1; }

CURRENT=$(git tag -l 'v*' | sed 's/^v//' | sort -V | tail -1)
[ -n "$CURRENT" ] || { echo "no tags found" >&2; exit 1; }
case "$CURRENT" in
  [0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "latest tag v$CURRENT is not major.minor.patch" >&2; exit 1 ;;
esac

IFS=. read -r MAJOR MINOR PATCH <<EOF
$CURRENT
EOF
case "$BUMP" in
  major) NEXT="$((MAJOR + 1)).0.0" ;;
  minor) NEXT="$MAJOR.$((MINOR + 1)).0" ;;
  patch) NEXT="$MAJOR.$MINOR.$((PATCH + 1))" ;;
esac

grep -q '^## \[Unreleased\]' CHANGELOG.md || {
  echo "CHANGELOG.md has no [Unreleased] section; document the release there first" >&2
  exit 1
}
[ -n "$(awk '/^## \[Unreleased\]/{f=1;next} /^## /{f=0} f && NF' CHANGELOG.md)" ] || {
  echo "CHANGELOG.md [Unreleased] section is empty" >&2
  exit 1
}

echo "==> v$CURRENT -> v$NEXT ($BUMP)"
make check

awk -v next="$NEXT" -v date="$(date +%F)" '
  !done && $0 == "## [Unreleased]" {
    print; print ""; print "## [" next "] - " date; done = 1; next
  }
  { print }
' CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md

git add CHANGELOG.md
git commit -q -m "release: v$NEXT"
git push -q origin dev dev:main
git tag "v$NEXT"
git push -q origin "v$NEXT"

NOTES=$(awk -v v="$NEXT" '
  index($0, "## [" v "]") == 1 { found = 1; next }
  /^## / { found = 0 }
  found
' CHANGELOG.md)
gh release create "v$NEXT" --title "v$NEXT" --notes "$NOTES"

echo "==> released v$NEXT"
