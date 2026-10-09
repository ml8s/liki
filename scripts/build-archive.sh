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

ARCHIVE="$DIST_DIR/$SKILL_NAME.tar.gz"
echo "[build-archive] 打包 $SKILL_NAME..."

# 根 skill（liki）是方法论卡的唯一权威源，已自包含各域方法论卡（references/natal/references/divination/
# references/fengshui/naming 的 domains/）。直接打包根 skill 即可，无需合并独立专家包。
# 独立专家包（expert-packs/，1b 起）由 make sync-expert-packs 从根生成（1c）。

# 可复现：固定排序/时间戳/属主，gzip -n 去除时间戳。
tar --sort=name --mtime='@0' --owner=0 --group=0 --numeric-owner \
    --transform 's|^\./||' \
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
    -cf - . | gzip -n > "$ARCHIVE"

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

# 根 skill 自包含各域方法论卡（references/natal/references/divination/references/fengshui/naming 的 domains/），
# 归档必须包含它们，保证对外全能力。
for domain in references/natal/domains references/divination/domains references/fengshui/domains references/naming/domains; do
    if ! grep -Fq "$domain/" <<<"$ARCHIVE_LISTING"; then
        echo "[build-archive] error: archive missing root domain methodology $domain/" >&2
        exit 1
    fi
done
# 根主 SKILL.md 必须存在（方法论的 SKILL.md 入口是子域文件，不计数）
if ! grep -Fxq 'SKILL.md' <<<"$ARCHIVE_LISTING"; then
    echo "[build-archive] error: archive missing root SKILL.md" >&2
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
