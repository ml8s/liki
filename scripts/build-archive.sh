#!/bin/bash
# Build the single unified Liki skill archive.
# 工程根 = liki（仓库根）；skills/liki = 唯一可安装 skill。
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_DIR/dist"
SKILL_NAME="liki"
SKILL_DIR="$PROJECT_DIR/skills/$SKILL_NAME"

mkdir -p "$DIST_DIR"

if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
    echo "[build-archive] error: $SKILL_DIR/SKILL.md not found" >&2
    exit 1
fi
if [ -e "$SKILL_DIR/VERSION" ]; then
    echo "[build-archive] error: stale $SKILL_DIR/VERSION found; use VERSION.txt" >&2
    exit 1
fi
if [ ! -f "$SKILL_DIR/VERSION.txt" ]; then
    echo "[build-archive] error: $SKILL_DIR/VERSION.txt not found" >&2
    exit 1
fi

# Every domain manifest stays domain-local, but its distributed version is
# injected from the single skill distribution VERSION.txt file.
VERSION="$(tr -d '\r\n' < "$SKILL_DIR/VERSION.txt")"
find "$SKILL_DIR" -mindepth 3 -maxdepth 3 -type f -name skill-tools.json -print0 |
while IFS= read -r -d '' manifest; do
    python3 - "$manifest" "$VERSION" <<'PYEOF'
import json, sys
path, version = sys.argv[1:]
d = json.load(open(path, encoding="utf-8"))
info = d.setdefault("info", {})
if info.get("version") != version:
    info["version"] = version
    json.dump(d, open(path, "w", encoding="utf-8"), ensure_ascii=False, indent=2)
    print(f"  ✓ {path} info.version → {version}")
else:
    print(f"  ✓ {path} info.version 已是最新（{version}）")
PYEOF
done

ARCHIVE="$DIST_DIR/$SKILL_NAME.tar.gz"
echo "[build-archive] 打包 $SKILL_NAME..."

# Natal orchestration cards intentionally reference expert methodology paths.
# Include those read-only cards so the unified archive is self-contained; the
# optional WorkBuddy expert plugins remain separate interactive agents.
EXPERT_BAZI_DIR="$PROJECT_DIR/skills/liki-bazi"
EXPERT_ZIWEI_DIR="$PROJECT_DIR/skills/liki-ziwei"
for required in "$EXPERT_BAZI_DIR/skills/bazi/career.md" \
                "$EXPERT_ZIWEI_DIR/skills/ziwei/yingqi.md"; do
    if [ ! -f "$required" ]; then
        echo "[build-archive] error: required methodology not found: $required" >&2
        exit 1
    fi
done

tar czf "$ARCHIVE" \
    --transform 's|^\./||' \
    --transform 's|^skills/bazi/|liki-bazi/skills/bazi/|' \
    --transform 's|^skills/ziwei/|liki-ziwei/skills/ziwei/|' \
    --exclude skills/bazi/SKILL.md \
    --exclude skills/ziwei/SKILL.md \
    -C "$SKILL_DIR" \
    --exclude .git \
    --exclude .github \
    --exclude .githooks \
    --exclude .claude \
    --exclude .reasonix \
    --exclude .pytest_cache \
    --exclude __pycache__ \
    --exclude CHANGELOG.md \
    --exclude '*.tar.gz' \
    --exclude dist \
    . \
    -C "$EXPERT_BAZI_DIR" skills/bazi \
    -C "$EXPERT_ZIWEI_DIR" skills/ziwei

DESC="$(sed -n 's/^description: //p' "$SKILL_DIR/SKILL.md" | head -1 | sed 's/^"//;s/"$//')"
echo "  ✓ $ARCHIVE ($(du -h "$ARCHIVE" | cut -f1))"

ARCHIVE_LISTING="$(tar -tzf "$ARCHIVE")"

if grep -Fxq 'VERSION' <<<"$ARCHIVE_LISTING"; then
    echo "[build-archive] error: archive contains unsupported extensionless VERSION" >&2
    exit 1
fi
if ! grep -Fxq 'VERSION.txt' <<<"$ARCHIVE_LISTING"; then
    echo "[build-archive] error: archive missing VERSION.txt" >&2
    exit 1
fi
if grep -E '\.(cmd|bat|ps1)$' <<<"$ARCHIVE_LISTING" | grep -q .; then
    echo "[build-archive] error: archive contains platform-specific launcher" >&2
    exit 1
fi
if grep -Fxq 'requirements.txt' <<<"$ARCHIVE_LISTING"; then
    echo "[build-archive] error: runtime Python dependencies must come from MCP services" >&2
    exit 1
fi
for required_method in \
    liki-bazi/skills/bazi/career.md \
    liki-ziwei/skills/ziwei/yingqi.md; do
    if ! grep -Fxq "$required_method" <<<"$ARCHIVE_LISTING"; then
        echo "[build-archive] error: archive missing $required_method" >&2
        exit 1
    fi
done
manifest_count="$(grep -Ec '(^|/)SKILL\.md$' <<<"$ARCHIVE_LISTING")"
if [ "$manifest_count" -ne 1 ]; then
    echo "[build-archive] error: expected exactly one SKILL.md, got $manifest_count" >&2
    exit 1
fi

INDEX="$DIST_DIR/index.json"
python3 - "$ARCHIVE" "$SKILL_DIR" "$SKILL_NAME" "$DESC" <<'PYEOF'
import hashlib, json, sys
archive, skill_dir, name, description = sys.argv[1:]
digest = "sha256:" + hashlib.sha256(open(archive, "rb").read()).hexdigest()
entry = {"name": name, "type": "archive", "url": f"/skills/{name}.tar.gz",
         "digest": digest, "description": description}
with open("dist/index.json", "w", encoding="utf-8") as fh:
    json.dump({"$schema": "https://schemas.agentskills.io/discovery/0.2.0/schema.json",
               "skills": [entry]}, fh, ensure_ascii=False, indent=2)
PYEOF
echo "  ✓ $INDEX"
