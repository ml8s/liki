#!/bin/bash
# Build skill archive (liki.tar.gz + index.json) into dist/
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_DIR/dist"

ARCHIVE="$DIST_DIR/liki.tar.gz"
INDEX="$DIST_DIR/index.json"

mkdir -p "$DIST_DIR"

echo "[build-archive] 打包 skills..."
tar czf "$ARCHIVE" \
    --transform 's|^\./||' \
    -C "$PROJECT_DIR" \
    --exclude .git \
    --exclude .gitignore \
    --exclude .githooks \
    --exclude .claude \
    --exclude .reasonix \
    --exclude '*.tar.gz' \
    --exclude dist \
    --exclude reasonix.toml \
    --exclude README.md \
    --exclude README.en.md \
    --exclude CHANGELOG.md \
    --exclude CONTRIBUTING.md \
    --exclude LICENSE \
    --exclude AGENTS.md \
    --exclude Makefile \
    .

DIGEST="sha256:$(sha256sum "$ARCHIVE" | awk '{print $1}')"

cat > "$INDEX" <<EOF
{
  "\$schema": "https://schemas.agentskills.io/discovery/0.2.0/schema.json",
  "skills": [
    {
      "name": "liki",
      "type": "archive",
      "url": "/skills/liki.tar.gz",
      "digest": "$DIGEST",
      "description": "liki — 命理 AI Agent Skills，八字/紫微/起名/六爻/奇门/黄历/风水"
    }
  ]
}
EOF

echo "  ✓ $ARCHIVE ($(du -h "$ARCHIVE" | cut -f1))"
echo "  ✓ $INDEX ($DIGEST)"
