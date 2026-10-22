#!/usr/bin/env python3
"""WorkBuddy 制品版本投影：发行 CalVer → 平台 semver。

平台按 semver 解析制品版本，4 段 CalVer 会被直接拒（实测专家包返回
「版本号 "2026.10.16.2" 格式不合法，请使用 semver 格式（如 1.0.0）」），
而项目内部只有一个月/日/序号版本 `YYYY.MM.DD.SERIAL`。投影规则只在这里定义一次：

    YYYY.MM.DD.SERIAL  →  YYYY.MMDD.SERIAL        （去掉月日的前导零）

性质：

* 合法 semver——三段纯数字，无前导零，无预发布/构建元数据；
* 严格递增——同日内靠 SERIAL，跨日靠 MMDD（1/31 → 131 < 2/1 → 201），跨年靠 YYYY；
* 可还原——MINOR 恒为 3～4 位，左侧补零即得 `MMDD`（101 → 0101 = 1 月 1 日）。

制品侧版本（技能 `SKILL.md`、专家包身份/域 skill/`plugin.json`、连接器
`connector-meta.json`）一律由 `workbuddy_version()` 从 `skills/liki/VERSION.txt`
派生：不手写、不新增计数源，bump 与校验走同一条规则，所以不会漂移。
"""

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
RUNTIME_VERSION_PATH = ROOT / "skills" / "liki" / "VERSION.txt"

CALVER_RE = re.compile(r"^(\d{4})\.(\d{2})\.(\d{2})\.(\d+)$")


def runtime_calver(path: Path = RUNTIME_VERSION_PATH) -> str:
    """发行 CalVer（唯一的版本真相源）。"""
    return path.read_text(encoding="utf-8").strip()


def workbuddy_version(calver: str) -> str:
    """把发行 CalVer 投影成平台 semver；输入非法即抛错（fail-closed）。"""
    match = CALVER_RE.match(calver.strip())
    if match is None:
        raise ValueError(f"发行版本必须是 YYYY.MM.DD.SERIAL，当前 {calver!r}")
    year, month, day, serial = match.groups()
    return f"{year}.{int(month + day)}.{int(serial)}"


def main() -> int:
    try:
        version = workbuddy_version(runtime_calver())
    except ValueError as exc:
        print(f"[version-projection] error: {exc}", file=sys.stderr)
        return 1
    print(version)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
