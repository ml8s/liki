#!/usr/bin/env bash
# Build, smoke, and push the liki release images for a CalVer release tag.
#
# Invoked by release.yml on a pushed vYYYY.MM.DD.N tag. Requires a prior
# `make release-gate` (provides dist/ for the experts assembly image) and a
# logged-in GHCR session. Emits digest outputs for the workflow.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RELEASE_VERSION="${RELEASE_VERSION:?set RELEASE_VERSION to the CalVer version without the v prefix (e.g. 2026.10.16.0)}"
if ! [[ "$RELEASE_VERSION" =~ ^[0-9]{4}\.[0-9]{2}\.[0-9]{2}\.[0-9]+$ ]]; then
  echo "❌ RELEASE_VERSION must be CalVer YYYY.MM.DD.N: $RELEASE_VERSION" >&2
  exit 1
fi
RELEASE_TAG="v$RELEASE_VERSION"
SOURCE_SHA="$(git rev-parse HEAD)"
TAG_SHA="$(git rev-parse --verify "refs/tags/$RELEASE_TAG^{commit}" 2>/dev/null || true)"
if [ "$TAG_SHA" != "$SOURCE_SHA" ]; then
  echo "❌ checked-out commit is not the $RELEASE_TAG tag ($SOURCE_SHA)" >&2
  exit 1
fi

BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
OUT_DIR="$ROOT/build/release"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

[ -d dist/agents ] || make build-deployment

AGENTS_BASE="$(python3 scripts/agent_runtime_pin.py --ref)"
REGISTRY="ghcr.io/ml8s"

build_image() {
  local metadata_name="$1" image_name="$2" context="$3" dockerfile="$4"
  shift 4
  echo "── building $REGISTRY/$image_name:$RELEASE_TAG ──"
  docker buildx build \
    --file "$dockerfile" \
    --metadata-file "$OUT_DIR/$metadata_name.json" \
    --provenance=mode=max \
    --sbom=true \
    --label "org.opencontainers.image.revision=$SOURCE_SHA" \
    --label "org.opencontainers.image.version=$RELEASE_TAG" \
    --build-arg "VERSION=$RELEASE_VERSION" \
    --build-arg "COMMIT=$SOURCE_SHA" \
    --build-arg "BUILD_TIME=$BUILD_TIME" \
    --tag "$REGISTRY/$image_name:$RELEASE_TAG" \
    --tag "$REGISTRY/$image_name:latest" \
    --push \
    "$@" \
    "$context"
}

build_image engine liki-engine engine engine/dev/Dockerfile
build_image counsel liki-counsel . counsel/Dockerfile

# The experts image assembles dist/agents + skills + expert-packs on the
# pinned liki-agent-runtime base (mirrors make image-assembly).
rm -rf assembly/dist
mkdir -p assembly/dist/agents
cp -r dist/agents/. assembly/dist/agents/
cp -r skills assembly/dist/skills
cp -r expert-packs assembly/dist/expert-packs
build_image experts liki-experts assembly assembly/Dockerfile \
  --build-arg "BASE_IMAGE=$AGENTS_BASE"
rm -rf assembly/dist

digest_of() {
  python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["containerimage.digest"])' \
    "$OUT_DIR/$1.json"
}

ENGINE_DIGEST="$(digest_of engine)"
COUNSEL_DIGEST="$(digest_of counsel)"
EXPERTS_DIGEST="$(digest_of experts)"

echo "── release smoke ──"
ENGINE_IMAGE="$REGISTRY/liki-engine@$ENGINE_DIGEST" \
COUNSEL_IMAGE="$REGISTRY/liki-counsel@$COUNSEL_DIGEST" \
IMAGE_TAG=unused \
  scripts/smoke-image.sh

for profile in natal divination fengshui naming; do
  docker run --rm "$REGISTRY/liki-experts@$EXPERTS_DIGEST" \
    validate -deployment "/deployment/$profile/deployment.json"
done

python3 scripts/build_candidate_manifest.py \
  --source-commit "$SOURCE_SHA" \
  --candidate-tag "$RELEASE_TAG" \
  --registry "ghcr.io" \
  --metadata-dir "$OUT_DIR" \
  --output "$OUT_DIR/candidate-digests.json"

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    echo "engine_digest=$ENGINE_DIGEST"
    echo "counsel_digest=$COUNSEL_DIGEST"
    echo "experts_digest=$EXPERTS_DIGEST"
    echo "manifest=$OUT_DIR/candidate-digests.json"
  } >> "$GITHUB_OUTPUT"
fi

echo "✅ release images pushed: $RELEASE_TAG"
