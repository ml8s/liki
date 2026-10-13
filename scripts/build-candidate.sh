#!/usr/bin/env bash
# Build and publish a complete local release candidate from a clean tree.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

require_clean_worktree() {
  local dirty=0
  git diff --quiet || dirty=1
  git diff --cached --quiet || dirty=1
  if [ "$dirty" -ne 0 ] || [ -n "$(git status --porcelain)" ]; then
    echo "✗ candidate requires a clean Git worktree" >&2
    git status --short >&2
    exit 1
  fi
}

require_clean_worktree
RELEASE_VERSION="${RELEASE_VERSION:?set RELEASE_VERSION to the selected CalVer, e.g. 2026.10.13.0}"
if [[ ! "$RELEASE_VERSION" =~ ^[0-9]{4}\.[0-9]{2}\.[0-9]{2}\.[0-9]+$ ]]; then
  echo "✗ RELEASE_VERSION must use CalVer YYYY.MM.DD.N: $RELEASE_VERSION" >&2
  exit 1
fi
version_file="$(cat skills/liki/VERSION.txt)"
if [ "$version_file" != "$RELEASE_VERSION" ]; then
  echo "✗ RELEASE_VERSION=$RELEASE_VERSION does not match skills/liki/VERSION.txt=$version_file" >&2
  exit 1
fi


make gate


if [ -n "$(git status --porcelain)" ]; then
  echo "❌ Git worktree changed during gate; aborting candidate build" >&2
  git status --short >&2
  exit 1
fi
SOURCE_SHA="$(git rev-parse HEAD)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
CANDIDATE_TAG="candidate-${SOURCE_SHA:0:7}-$(date -u +%Y%m%d%H%M%S)"
REGISTRY="ghcr.io/ml8s"
AGENTS_BASE="$(python3 scripts/liki_agents_pin.py --ref)"
METADATA_DIR="$(mktemp -d)"
trap 'rm -rf "$METADATA_DIR"' EXIT

build_image() {
  local metadata_name="$1" image_name="$2" context="$3" dockerfile="$4"; shift 4
  local tag="$REGISTRY/$image_name:$CANDIDATE_TAG"
  echo "── building and pushing $tag ──"
  docker buildx build \
    --file "$dockerfile" \
    --metadata-file "$METADATA_DIR/$metadata_name.json" \
    --provenance=mode=max \
    --sbom=true \
    --push \
    --tag "$tag" \
    "$@" \
    "$context"
}

build_image engine liki-engine engine/dev/Dockerfile \
  --build-arg VERSION="$RELEASE_VERSION" \
  --build-arg COMMIT="$SOURCE_SHA" \
  --build-arg BUILD_TIME="$BUILD_TIME"
build_image counsel liki-counsel . counsel/Dockerfile \
  --build-arg VERSION="$RELEASE_VERSION" \
  --build-arg COMMIT="$SOURCE_SHA" \
  --build-arg BUILD_TIME="$BUILD_TIME"

rm -rf assembly/dist
mkdir -p assembly/dist/agents
cp -r dist/agents/. assembly/dist/agents/
cp -r skills assembly/dist/skills
cp -r expert-packs assembly/dist/expert-packs
build_image experts liki-multi-expert assembly assembly/Dockerfile \
  --build-arg BASE_IMAGE="$AGENTS_BASE" \
  --build-arg PROFILE=experts \
  --build-arg VERSION="$RELEASE_VERSION" \
  --build-arg COMMIT="$SOURCE_SHA" \
  --build-arg BUILD_TIME="$BUILD_TIME"
build_image single_expert liki-single-expert assembly assembly/Dockerfile \
  --build-arg BASE_IMAGE="$AGENTS_BASE" \
  --build-arg PROFILE=single \
  --build-arg VERSION="$RELEASE_VERSION" \
  --build-arg COMMIT="$SOURCE_SHA" \
  --build-arg BUILD_TIME="$BUILD_TIME"
rm -rf assembly/dist

metadata_digest() {
  python3 - "$METADATA_DIR/$1.json" <<'PY'
import json, sys
data = json.load(open(sys.argv[1], encoding="utf-8"))
print(data["containerimage.digest"])
PY
}

# Existing smoke support accepts image references via these variables.
ENGINE_IMAGE="$REGISTRY/liki-engine@$(metadata_digest engine)" \
  COUNSEL_IMAGE="$REGISTRY/liki-counsel@$(metadata_digest counsel)" \
  IMAGE_TAG=unused scripts/smoke-image.sh

for profile in experts single_expert; do
  image_name="$([ "$profile" = experts ] && echo liki-multi-expert || echo liki-single-expert)"
  docker run --rm \
    "$REGISTRY/$image_name@$(metadata_digest "$profile")" \
    validate -deployment /deployment/deployment.json
done

python3 scripts/build_candidate_manifest.py \
  --source-commit "$SOURCE_SHA" \
  --candidate-tag "$CANDIDATE_TAG" \
  --metadata-dir "$METADATA_DIR" \
  --output build/candidate/candidate-digests.json

echo "✅ candidate published: $CANDIDATE_TAG"
cat build/candidate/candidate-digests.json
