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
    --exclude webapp \
    --exclude tests \
    --exclude VERSION \
    --exclude scripts \
    .

DIGEST="sha256:$(sha256sum "$ARCHIVE" | awk '{print $1}')"

# 内容指纹（外部评审 #1：版本同内容滞后自检须能发现）——随 archive 分发 content.sha256
python3 "$PROJECT_DIR/tools/hash.py" "$PROJECT_DIR" "$PROJECT_DIR/content.sha256"

# 从 SKILL.md frontmatter 读取 description（单一事实源，避免硬编码漂移）
DESC="$(sed -n 's/^description: //p' "$SCRIPT_DIR/../SKILL.md" | head -1)"

cat > "$INDEX" <<EOF
{
  "\$schema": "https://schemas.agentskills.io/discovery/0.2.0/schema.json",
  "skills": [
    {
      "name": "liki",
      "type": "archive",
      "url": "/skills/liki.tar.gz",
      "digest": "$DIGEST",
      "description": "$DESC"
    }
  ]
}
EOF

echo "  ✓ $ARCHIVE ($(du -h "$ARCHIVE" | cut -f1))"
echo "  ✓ $INDEX ($DIGEST)"
