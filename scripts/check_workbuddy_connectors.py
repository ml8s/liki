#!/usr/bin/env python3
"""校验 `workbuddy/` 下连接器提交包是否符合 WorkBuddy 连接器规范。

规范来源：https://open.workbuddy.cn/docs/connector 的「连接器元信息」「mcp.json」「图标规范」
「提交前检查」四节。检查项只覆盖可机械判定、且在提审前必须为真的部分：

* 包结构：`connector-meta.json`（必须）+ `mcp.json` 或 `cli.json`（二选一）+ `icon.*`；
* 元信息：必填字段、描述字数、`source` kebab-case 且全局唯一、示例条数与类型；
* MCP：`mcpServers` 恰好一个 Server，远程地址 HTTPS，超时为正整数；
* 版本门：使用某字段声明的 `minWorkbuddyVersion` 不低于该字段的最低版本；
* 凭证：包内不出现真实凭证（占位符一律 `${VAR}` 形式）。

只读，不生成工件；由 `make check` 调用。
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import runtime_calver, workbuddy_version  # noqa: E402

WORKBUDDY = ROOT / "workbuddy"
META_NAME = "connector-meta.json"
CONFIG_NAMES = ("mcp.json", "cli.json")
ICON_NAMES = ("icon.svg", "icon.png", "icon.jpg")
# 图标与专家头像同族：字形必须转路径（宿主字体与 dominant-baseline 都不可控）。
ICON_TEXT_MARKERS = ("<text", "font-family", "dominant-baseline")
TOKEN_SCHEMA_NAME = "token-schema.json"

NAME_REQUIRED = ("name", "name_en")
DESC_FIELDS = ("description", "description_zh", "description_en")
EXAMPLES_FIELDS = ("examples_zh", "examples_en")
SOURCES_REQUIRED = ("source",)
META_REQUIRED_STRINGS = NAME_REQUIRED + DESC_FIELDS + SOURCES_REQUIRED

TYPES = ("mcp", "cli", "skill-only")
AUTH_MODES = ("server-side", "gateway", "token")
TRANSPORTS = ("sse", "streamableHttp", "stdio")

SOURCE_RE = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")
VERSION_RE = re.compile(r"^(\d+)\.(\d+)\.(\d+)(?:[-+][0-9A-Za-z.-]+)?$")
PLACEHOLDER_RE = re.compile(r"^\$\{[A-Za-z0-9_]+\}$")
PLACEHOLDER_SCAN_RE = re.compile(r"\$\{([A-Za-z0-9_]+)\}")
SECRET_KEY_RE = re.compile(r"(?i)(secret|token|password|passwd|api[-_]?key|access[-_]?key|private[-_]?key)")

DESC_MIN, DESC_MAX = 20, 100
EXAMPLES_MIN, EXAMPLES_MAX = 2, 5
REMOTE_TYPES = ("sse", "streamableHttp")

# 字段 → 最低 WorkBuddy 版本（「版本兼容性」表）。
FIELD_MIN_VERSIONS: tuple[tuple[tuple[str, ...], tuple[int, int, int]], ...] = (
    (("name_zh", "name_en", *EXAMPLES_FIELDS, "statusMatchJson", "versionCheck"), (4, 24, 0)),
    (("minWorkbuddyVersion", "maxWorkbuddyVersion"), (4, 22, 12)),
    (("name_map", "description_map"), (5, 2, 0)),
)
CONFIG_MIN_VERSIONS: tuple[tuple[str, tuple[int, int, int]], ...] = (
    ("disabledTools", (4, 22, 15)),
    ("cwd", (4, 22, 15)),
    ("preAuth", (5, 0, 0)),
    ("runtime", (5, 0, 0)),
    ("staticEnv", (5, 0, 0)),
    ("staticHeaders", (5, 0, 0)),
)


def parse_version(value: str) -> tuple[int, int, int] | None:
    match = VERSION_RE.match(value)
    if not match:
        return None
    return tuple(int(part) for part in match.groups()[:3])  # type: ignore[return-value]


def load_json(path: Path, errors: list[str]) -> dict | None:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        errors.append(f"{path.relative_to(ROOT)}: JSON 解析失败（{exc}）")
        return None
    if not isinstance(data, dict):
        errors.append(f"{path.relative_to(ROOT)}: 顶层必须是对象")
        return None
    return data


def check_credentials(path: Path, data: object, errors: list[str]) -> None:
    """键名像凭证且值不是 ${VAR} 占位符，即判为疑似真实凭证。"""
    rel = path.relative_to(ROOT)
    stack: list[tuple[str, object]] = [("", data)]
    while stack:
        key, node = stack.pop()
        if isinstance(node, dict):
            for child_key, child in node.items():
                if isinstance(child, str) and SECRET_KEY_RE.search(str(child_key)):
                    if child.strip() and not PLACEHOLDER_RE.match(child.strip()):
                        errors.append(
                            f"{rel}: 疑似真实凭证 `{child_key}`，应写成 ${{VAR}} 占位符"
                        )
                stack.append((str(child_key), child))
        elif isinstance(node, list):
            for item in node:
                stack.append((key, item))
        elif isinstance(node, str) and "Bearer " in node and not PLACEHOLDER_SCAN_RE.search(node):
            errors.append(f"{rel}: `{key}` 含字面量 Bearer 凭证，应写成 `Bearer ${{VAR}}`")


def check_by_key_version_gate(
    fields: tuple[tuple[tuple[str, ...], tuple[int, int, int]], ...],
    data: dict,
    declared: tuple[int, int, int] | None,
    rel: Path,
    label: str,
    errors: list[str],
) -> None:
    for keys, minimum in fields:
        if not any(key in data for key in keys):
            continue
        if declared is None:
            errors.append(f"{rel}: 使用 `{label}` 字段时必须声明 minWorkbuddyVersion")
        elif declared < minimum:
            errors.append(
                f"{rel}: `{label}` 需要 WorkBuddy >= "
                f"{'.'.join(str(part) for part in minimum)}，当前声明过低"
            )


def check_mcp_config(path: Path, data: dict, meta: dict, errors: list[str]) -> None:
    rel = path.relative_to(ROOT)
    servers = data.get("mcpServers")
    if not isinstance(servers, dict) or not servers:
        errors.append(f"{rel}: 缺少 `mcpServers` 或为空")
        return
    if len(servers) != 1:
        errors.append(f"{rel}: 一个连接器只能声明一个 MCP Server，当前 {len(servers)} 个")

    placeholders: set[str] = set()
    for name, server in servers.items():
        if not isinstance(server, dict):
            errors.append(f"{rel}: server `{name}` 必须是对象")
            continue
        transport = server.get("type", "streamableHttp")
        if transport not in TRANSPORTS:
            errors.append(f"{rel}: server `{name}` 的 type `{transport}` 不在 {TRANSPORTS}")
        if transport in REMOTE_TYPES:
            url = server.get("url")
            if not isinstance(url, str) or not url.startswith("https://"):
                errors.append(f"{rel}: server `{name}` 远程地址必须使用 HTTPS")
        elif transport == "stdio" and not server.get("command"):
            errors.append(f"{rel}: server `{name}` stdio 传输缺少 `command`")
        timeout = server.get("timeout")
        if timeout is not None and (not isinstance(timeout, int) or timeout <= 0):
            errors.append(f"{rel}: server `{name}` 的 timeout 必须是正整数")
        placeholders |= set(PLACEHOLDER_SCAN_RE.findall(json.dumps(server, ensure_ascii=False)))
        check_by_key_version_gate(
            CONFIG_MIN_VERSIONS, server, parse_version(str(meta.get("minWorkbuddyVersion", ""))),
            rel, f"mcp.json.{name}", errors,
        )

    if meta.get("auth_mode") != "token":
        return
    schema_path = path.parent / TOKEN_SCHEMA_NAME
    if not schema_path.is_file():
        errors.append(f"{rel}: auth_mode 为 token 时必须提供 {TOKEN_SCHEMA_NAME}")
        return
    schema = load_json(schema_path, errors)
    if schema is None:
        return
    fields = schema.get("fields")
    if not isinstance(fields, list) or not fields:
        errors.append(f"{TOKEN_SCHEMA_NAME}: `fields` 至少一项")
        return
    keys = {str(field.get("key")) for field in fields if isinstance(field, dict)}
    if keys != placeholders:
        errors.append(
            f"{TOKEN_SCHEMA_NAME}: 表单字段 key {sorted(keys)} 与 mcp.json 占位符 "
            f"{sorted(placeholders)} 必须一一对应"
        )


def check_meta(meta: dict, directory: Path, icon: Path | None, configs: list[str], errors: list[str]) -> None:
    rel = directory.relative_to(ROOT)
    for key in META_REQUIRED_STRINGS:
        value = meta.get(key)
        if not isinstance(value, str) or not value.strip():
            errors.append(f"{rel}: 缺少必填字符串字段 `{key}`")
    for key in DESC_FIELDS:
        value = meta.get(key)
        if isinstance(value, str) and value.strip():
            length = len(value.strip())
            if not DESC_MIN <= length <= DESC_MAX:
                errors.append(f"{rel}: `{key}` 建议 {DESC_MIN}～{DESC_MAX} 字，当前 {length} 字")
    for key in EXAMPLES_FIELDS:
        value = meta.get(key)
        if not isinstance(value, list):
            errors.append(f"{rel}: `{key}` 必须是数组")
            continue
        if not EXAMPLES_MIN <= len(value) <= EXAMPLES_MAX:
            errors.append(
                f"{rel}: `{key}` 建议 {EXAMPLES_MIN}～{EXAMPLES_MAX} 条，当前 {len(value)} 条"
            )
        for item in value:
            if not isinstance(item, str) or not item.strip():
                errors.append(f"{rel}: `{key}` 的示例必须是非空字符串")

    source = meta.get("source")
    if isinstance(source, str) and not SOURCE_RE.match(source):
        errors.append(f"{rel}: `source` 只能用小写字母、数字和连字符，当前 `{source}`")

    meta_type = meta.get("type", "mcp")
    if meta_type not in TYPES:
        errors.append(f"{rel}: `type` 必须是 {TYPES} 之一，当前 `{meta_type}`")
    expected = "cli.json" if meta_type == "cli" else "mcp.json"
    if configs and expected not in configs:
        errors.append(f"{rel}: type 为 `{meta_type}` 时配置应为 {expected}，当前 {configs}")

    version = meta.get("version")
    expected_version = workbuddy_version(runtime_calver())
    if version is None:
        errors.append(f"{rel}: 缺少 `version`（制品面字段，取值见 scripts/version_projection.py）")
    elif parse_version(str(version)) is None:
        errors.append(f"{rel}: `version` 必须是语义化版本号，当前 `{version}`")
    elif str(version) != expected_version:
        errors.append(
            f"{rel}: `version` 必须是发行版本 {runtime_calver()} 的平台投影 `{expected_version}`，"
            f"当前 `{version}`；由 scripts/bump_version.py 更新，不手写"
        )

    declared = parse_version(str(meta.get("minWorkbuddyVersion", "")))
    if "minWorkbuddyVersion" in meta and declared is None:
        errors.append(f"{rel}: `minWorkbuddyVersion` 必须是语义化版本号")
    if declared is None:
        errors.append(f"{rel}: 使用 name_en / examples_* 时必须声明 `minWorkbuddyVersion`")
    else:
        check_by_key_version_gate(FIELD_MIN_VERSIONS, meta, declared, rel, "connector-meta", errors)

    auth_mode = meta.get("auth_mode")
    if auth_mode is not None and auth_mode not in AUTH_MODES:
        errors.append(f"{rel}: `auth_mode` 必须是 {AUTH_MODES} 之一，当前 `{auth_mode}`")
    if auth_mode != "token":
        stray = directory / TOKEN_SCHEMA_NAME
        if stray.is_file():
            errors.append(f"{rel}: 未使用 auth_mode: token 时不应存在 {TOKEN_SCHEMA_NAME}")

    if icon is None:
        errors.append(f"{rel}: 缺少图标（{'、'.join(ICON_NAMES)} 之一）")


def iter_connectors() -> tuple[list[Path], list[str]]:
    errors: list[str] = []
    if not WORKBUDDY.is_dir():
        return [], [f"缺少 {WORKBUDDY.relative_to(ROOT)} 目录"]
    connectors: list[Path] = []
    for entry in sorted(WORKBUDDY.iterdir()):
        if entry.name == "README.md":
            continue
        if not entry.is_dir():
            errors.append(
                f"{WORKBUDDY.relative_to(ROOT)}/{entry.name}: 只允许 README.md 与连接器目录"
                "（提交用 zip 属产物，用 `make workbuddy-package` 生成到 dist/workbuddy/）"
            )
            continue
        if (entry / META_NAME).is_file():
            connectors.append(entry)
        else:
            errors.append(f"{entry.relative_to(ROOT)}: 缺少 {META_NAME}")
    return connectors, errors


def check_connector(directory: Path, sources: dict[str, Path]) -> list[str]:
    errors: list[str] = []
    rel = directory.relative_to(ROOT)
    meta = load_json(directory / META_NAME, errors)
    if meta is None:
        return errors
    check_credentials(directory / META_NAME, meta, errors)

    configs = [name for name in CONFIG_NAMES if (directory / name).is_file()]
    if len(configs) != 1:
        errors.append(f"{rel}: 必须且只能有 {'、'.join(CONFIG_NAMES)} 之一，当前 {configs}")

    source = meta.get("source")
    if isinstance(source, str) and source:
        if source in sources:
            errors.append(f"{rel}: `source` `{source}` 与 {sources[source].relative_to(ROOT)} 重复")
        else:
            sources[source] = directory

    icons = [directory / name for name in ICON_NAMES if (directory / name).is_file()]
    icon = icons[0] if icons else None
    if len(icons) > 1:
        errors.append(f"{rel}: 图标只应有 {'、'.join(ICON_NAMES)} 之一")
    if icon is not None:
        if icon.stat().st_size == 0:
            errors.append(f"{rel}: 图标文件为空")
        elif icon.suffix == ".svg":
            svg = icon.read_text(encoding="utf-8")
            if "<svg" not in svg:
                errors.append(f"{rel}: {icon.name} 不含 <svg> 根元素")
            for marker in ICON_TEXT_MARKERS:
                if marker in svg:
                    errors.append(f"{rel}: {icon.name} 图标只允许路径字形，出现 `{marker}`")

    if "mcp.json" in configs:
        config = load_json(directory / "mcp.json", errors)
        if config is not None:
            check_credentials(directory / "mcp.json", config, errors)
            check_mcp_config(directory / "mcp.json", config, meta, errors)
    elif "cli.json" in configs:
        config = load_json(directory / "cli.json", errors)
        if config is not None:
            check_credentials(directory / "cli.json", config, errors)

    check_meta(meta, directory, icon, configs, errors)
    return errors


def main() -> int:
    connectors, errors = iter_connectors()
    sources: dict[str, Path] = {}
    for directory in connectors:
        errors.extend(check_connector(directory, sources))

    if errors:
        for line in errors:
            print(f"[check-workbuddy] ✗ {line}", file=sys.stderr)
        print(f"[check-workbuddy] {len(errors)} 个问题", file=sys.stderr)
        return 1
    print(f"[check-workbuddy] ✓ {len(connectors)} 个连接器符合提交前检查")
    return 0


if __name__ == "__main__":
    sys.exit(main())
