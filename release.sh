#!/usr/bin/env bash
set -euo pipefail

# Usage: ./release.sh v0.2.0
# Idempotent — safe to re-run for the same version.

VERSION="${1:?usage: ./release.sh <version, e.g. v0.2.0>}"
REPO="Patrik-Stas/jj"
TAP_FORMULA="../patrik-homebrew-tap/Formula/jj.rb"

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "error: version must match vX.Y.Z (got: $VERSION)" >&2
    exit 1
fi

if [[ ! -f "$TAP_FORMULA" ]]; then
    echo "error: tap formula not found at $TAP_FORMULA" >&2
    exit 1
fi

DIST="dist/${VERSION}"
PLATFORMS=(darwin-arm64 darwin-amd64 linux-amd64 linux-arm64)
BINARIES=()
MISSING=false
for p in "${PLATFORMS[@]}"; do
    BINARIES+=("${DIST}/_jj-${p}")
    [[ -f "${DIST}/_jj-${p}" ]] || MISSING=true
done

if $MISSING; then
    echo "==> Cross-compiling $VERSION"
    make release VERSION="$VERSION"
else
    echo "==> Binaries already exist in ${DIST}/, skipping build"
fi

echo "==> Tagging $VERSION"
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "    tag $VERSION already exists, skipping"
else
    git tag "$VERSION"
    git push origin "$VERSION"
fi

echo "==> Creating GitHub release"
if gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1; then
    echo "    release $VERSION already exists, uploading binaries (overwriting if present)"
    gh release upload "$VERSION" "${BINARIES[@]}" --repo "$REPO" --clobber
else
    gh release create "$VERSION" "${BINARIES[@]}" --repo "$REPO" \
        --title "$VERSION" --notes "Release $VERSION"
fi

echo "==> Computing SHA256 hashes"
declare -A SHAS
for bin in "${BINARIES[@]}"; do
    SHAS[$bin]="$(shasum -a 256 "$bin" | cut -d' ' -f1)"
    echo "    $bin: ${SHAS[$bin]}"
done

echo "==> Updating tap formula"
SEMVER="${VERSION#v}"
sed -i.bak "s/version \".*\"/version \"$SEMVER\"/" "$TAP_FORMULA"
for p in "${PLATFORMS[@]}"; do
    sha="${SHAS[${DIST}/_jj-${p}]}"
    # Match the sha256 line that follows the url containing this platform
    sed -i.bak "/_jj-${p}/{n;s/sha256 \".*\"/sha256 \"${sha}\"/;}" "$TAP_FORMULA"
done
rm -f "${TAP_FORMULA}.bak"

echo "==> Tap formula updated. Don't forget to commit and push patrik-homebrew-tap."
echo ""
echo "    cd ../patrik-homebrew-tap"
echo "    git add Formula/jj.rb && git commit -m \"jj $VERSION\" && git push"
