#!/usr/bin/env python3
"""content.sha256 指纹工具（外部评审 #1：自检需能发现"版本同、内容滞后"）。

tree_sha256(dir) 计算 skill 文件树的稳定指纹（排除 dist/.git/__pycache__/tests 等非运行时文件）。
发布时由 build-archive.sh 生成 content.sha256 随 archive 分发；SKILL.md 自检对比
"本地 VERSION+content.sha256" 与远程，任一不一致即提示更新。
"""
from __future__ import annotations

import hashlib
import os

# 指纹不包含的目录/文件（非运行时内容：构建产物/版本控制/工程文件/测试/缓存/webapp）
# 与 scripts/build-archive.sh 打包范围保持一致——指纹只覆盖随包分发的内容
# （SKILL.md + app/ + domains/ + tools/ + VERSION + content.sha256）。
EXCLUDE_DIRS = {".git", "dist", "__pycache__", ".pytest_cache", ".claude", ".reasonix",
                "tests", "scripts", "webapp", ".github", ".githooks", "docs"}
EXCLUDE_FILES = {"content.sha256", "liki.tar.gz", "index.json",
                 "README.md", "README.en.md", "CHANGELOG.md", "CONTRIBUTING.md",
                 "LICENSE", "AGENTS.md", "Makefile", "pytest.ini", ".gitignore"}


def tree_sha256(root: str) -> str:
    """按相对路径稳定遍历（排序）计算文件树 SHA-256。

    只纳入运行时文件（排除 EXCLUDE_DIRS/EXCLUDE_FILES）——内容变化（哪怕 VERSION 不变）指纹必变。
    EXCLUDE_FILES 按「根级相对路径」精确匹配（如 "README.md" 只排除根 README.md，
    不误伤 app/README.md 等分发内容）——与 build-archive.sh 的 --exclude './README.md' 口径一致。
    """
    h = hashlib.sha256()
    rels = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in EXCLUDE_DIRS]
        for fn in filenames:
            rel = os.path.relpath(os.path.join(dirpath, fn), root)
            if rel in EXCLUDE_FILES:
                continue
            rels.append(rel)
    for rel in sorted(rels):
        with open(os.path.join(root, rel), "rb") as f:
            h.update(rel.encode())
            h.update(b"\x00")
            h.update(f.read())
            h.update(b"\x00")
    return h.hexdigest()


def main() -> int:
    """用法: python3 tools/hash.py <skill_root> [output_file]"""
    import sys
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    digest = tree_sha256(root)
    if len(sys.argv) > 2:
        with open(sys.argv[2], "w", encoding="utf-8") as f:
            f.write(digest + "\n")
        print(f"content.sha256 写入 {sys.argv[2]}: {digest[:16]}…")
    else:
        print(digest)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
