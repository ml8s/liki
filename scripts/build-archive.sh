#!/bin/bash
# Build skill archives (liki-<name>.tar.gz + index.json) into dist/
# 工程根 = liki-skills（仓库根）；skills/ 下每个子目录 = 一个独立 skill（唯一被安装的部分）
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_DIR/dist"   # 构建产物属工程区（仓库根），不进 skill 内容区

# 4 个 skill：命理 / 问卦 / 风水 / 起名
SKILLS=(liki-bazi liki-divination liki-fengshui liki-naming)

mkdir -p "$DIST_DIR"

SKILL_ENTRIES=""

for NAME in "${SKILLS[@]}"; do
    SKILL_DIR="$PROJECT_DIR/skills/$NAME"
    if [ ! -f "$SKILL_DIR/SKILL.md" ]; then
        echo "[build-archive] 警告：$SKILL_DIR 无 SKILL.md，跳过" >&2
        continue
    fi
    ARCHIVE="$DIST_DIR/$NAME.tar.gz"

    echo "[build-archive] 打包 $NAME..."
    tar czf "$ARCHIVE" \
        --transform 's|^\./||' \
        -C "$SKILL_DIR" \
        --exclude .git \
        --exclude .github \
        --exclude .githooks \
        --exclude .claude \
        --exclude .reasonix \
        --exclude .pytest_cache \
        --exclude __pycache__ \
        --exclude '*.tar.gz' \
        --exclude dist \
        .

    DIGEST="sha256:$(sha256sum "$ARCHIVE" | awk '{print $1}')"

    # 内容指纹（外部评审 #1：版本同内容滞后自检须能发现）——随 archive 分发 content.sha256
    python3 "$SKILL_DIR/tools/hash.py" "$SKILL_DIR" "$SKILL_DIR/content.sha256"

    # 从 SKILL.md frontmatter 读取 description（单一事实源，避免硬编码漂移）
    DESC="$(sed -n 's/^description: //p' "$SKILL_DIR/SKILL.md" | head -1)"

    echo "  ✓ $ARCHIVE ($(du -h "$ARCHIVE" | cut -f1))"
done

INDEX="$DIST_DIR/index.json"
python3 - "$PROJECT_DIR" <<'PYEOF'
import glob, hashlib, json, os, re, sys
project_dir = sys.argv[1]
entries = []
for arc in sorted(glob.glob(os.path.join(project_dir, "dist", "*.tar.gz"))):
    name = os.path.basename(arc)[:-len(".tar.gz")]
    skill_dir = os.path.join(project_dir, "skills", name)
    sk = os.path.join(skill_dir, "SKILL.md")
    if not os.path.exists(sk):
        continue
    desc = ""
    for ln in open(sk, encoding="utf-8"):
        if ln.startswith("description:"):
            desc = ln[len("description:"):].strip()
            break
    digest = "sha256:" + hashlib.sha256(open(arc, "rb").read()).hexdigest()
    entries.append({"name": name, "type": "archive", "url": f"/skills/{name}.tar.gz",
                    "digest": digest, "description": desc})
with open(os.path.join(project_dir, "dist", "index.json"), "w", encoding="utf-8") as fh:
    json.dump({"$schema": "https://schemas.agentskills.io/discovery/0.2.0/schema.json", "skills": entries},
              fh, ensure_ascii=False, indent=2)
PYEOF
echo "  ✓ $INDEX"
