#!/usr/bin/env bash
# Build the immutable liki-web skill bundle and its digest sidecars. The
# client-download archive is also the server-side skill source; do not package
# the same Markdown tree twice.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_DIR/dist"
SKILL_DIR="$PROJECT_DIR/skills/liki"
BUNDLE="$DIST_DIR/liki-web-skill-bundle.tar.gz"

mkdir -p "$DIST_DIR"
for path in "$DIST_DIR/liki.tar.gz" "$DIST_DIR/index.json" "$SKILL_DIR/SKILL.md" "$SKILL_DIR/VERSION.txt"; do
  if [ ! -f "$path" ]; then
    echo "[web-skill-bundle] error: required artifact missing: $path" >&2
    exit 1
  fi
done

VERSION="$(tr -d '\r\n' < "$SKILL_DIR/VERSION.txt")"
RELEASE_TAG="${LIKI_RELEASE_TAG:-unreleased}"
SOURCE_COMMIT="${LIKI_SOURCE_COMMIT:-$(git -C "$PROJECT_DIR" rev-parse HEAD)}"
ARCHIVE_SHA256="sha256:$(sha256sum "$DIST_DIR/liki.tar.gz" | cut -d' ' -f1)"

cat > "$DIST_DIR/manifest.json" <<JSON
{
  "schema_version": "1",
  "artifact": "liki-web-skill-bundle.tar.gz",
  "release_tag": "$RELEASE_TAG",
  "runtime_version": "$VERSION",
  "source_commit": "$SOURCE_COMMIT",
  "archive_sha256": "$ARCHIVE_SHA256"
}
JSON

# Reproducible tar + gzip metadata makes the release asset digest meaningful.
tar --sort=name --mtime='@0' --owner=0 --group=0 --numeric-owner \
  -cf - \
  --exclude .git --exclude .github --exclude .githooks --exclude .pytest_cache \
  --exclude __pycache__ --exclude CHANGELOG.md \
  -C "$PROJECT_DIR" webapp \
  -C "$DIST_DIR" liki.tar.gz index.json manifest.json |
  gzip -n > "$BUNDLE"

BUNDLE_SHA256=$(sha256sum "$BUNDLE" | cut -d' ' -f1)
printf '%s  %s\n' "$BUNDLE_SHA256" "$(basename "$BUNDLE")" > "$DIST_DIR/liki-web-skill-bundle.sha256"
python3 - "$DIST_DIR/manifest.json" "$BUNDLE_SHA256" "$DIST_DIR/liki-web-skill-bundle.manifest.json" <<'PY'
import json
import shutil
import sys
source, bundle_digest, destination = sys.argv[1:]
with open(source, encoding="utf-8") as fh:
    manifest = json.load(fh)
manifest["bundle_sha256"] = "sha256:" + bundle_digest
with open(destination, "w", encoding="utf-8") as fh:
    json.dump(manifest, fh, ensure_ascii=False, indent=2)
    fh.write("\n")
PY

echo "  ✓ $BUNDLE ($(du -h "$BUNDLE" | cut -f1))"
echo "  ✓ $DIST_DIR/liki-web-skill-bundle.sha256"
