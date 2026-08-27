#!/bin/bash
# Build skill archives (liki-<name>.tar.gz + index.json) into dist/
# 工程根 = liki（仓库根，与仓库名 ml8s/liki 一致）；skills/ 下每个子目录 = 一个独立 skill（唯一被安装的部分）
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

    # 版本号单一来源（外部评审 #22：skill-tools.json info.version 曾落后包版本）——
    # build 时从 VERSION 注入，杜绝多源漂移；无 skill-tools.json 的 skill（子流程三件套）跳过
    if [ -f "$SKILL_DIR/tools/skill-tools.json" ] && [ -f "$SKILL_DIR/VERSION" ]; then
        python3 - "$SKILL_DIR" <<'PYEOF'
import json, os, sys
skill_dir = sys.argv[1]
p = os.path.join(skill_dir, "tools", "skill-tools.json")
version = open(os.path.join(skill_dir, "VERSION"), encoding="utf-8").read().strip()
d = json.load(open(p, encoding="utf-8"))
info = d.setdefault("info", {})
if info.get("version") != version:
    info["version"] = version
    json.dump(d, open(p, "w", encoding="utf-8"), ensure_ascii=False, indent=2)
    print(f"  ✓ skill-tools.json info.version → {version}")
else:
    print(f"  ✓ skill-tools.json info.version 已是最新（{version}）")
PYEOF
    fi

    # 内容指纹（外部评审 #1/#14：版本同内容滞后自检须能发现）——必须先于打包重算，
    # 否则 archive 内 content.sha256 是上一代旧值，安装副本出现「声明指纹≠实际内容」的假同步
    echo "[build-archive] 重算 $NAME 内容指纹..."
    python3 "$SKILL_DIR/tools/hash.py" "$SKILL_DIR" "$SKILL_DIR/content.sha256"

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

    # 校验 archive 内一致性：archive 内 content.sha256 == archive 内实际内容指纹
    # （安装副本 = archive 解包结果，此项通过则安装后「本地指纹 == 本地实际内容」必然成立）
    TMP_EXTRACT="$(mktemp -d)"
    tar xzf "$ARCHIVE" -C "$TMP_EXTRACT"
    if [ -f "$TMP_EXTRACT/content.sha256" ]; then
        ARCHIVE_FP="$(cat "$TMP_EXTRACT/content.sha256")"
    else
        ARCHIVE_FP=""
    fi
    ACTUAL_FP="$(python3 "$TMP_EXTRACT/tools/hash.py" "$TMP_EXTRACT")"
    rm -rf "$TMP_EXTRACT"
    if [ "$ARCHIVE_FP" != "$ACTUAL_FP" ]; then
        echo "[build-archive] 错误：$NAME archive 内 content.sha256（$ARCHIVE_FP）≠ 实际内容指纹（$ACTUAL_FP），假同步未消除，请检查打包/指纹范围" >&2
        exit 1
    fi
    echo "  ✓ $NAME archive 指纹一致（${ACTUAL_FP:0:16}…）"

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
