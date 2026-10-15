#!/usr/bin/env python3
"""Resolve the engine's MCP tool names from the typed registry.

``engine/internal/mcpcontract`` is the only engine tool source: ``NewEngine()``
registers every tool as ``Name: "<domain>.<tool>"``. The schema check and the
contract tests resolve the tool set through here, so they cannot drift from each
other or pin a file name that may move.
"""
from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CONTRACT_DIR = ROOT / "engine" / "internal" / "mcpcontract"

_METHOD_NAME = re.compile(r'Name:\s*"([a-z]+(?:\.[a-z_]+)+)"')


def engine_method_names() -> set[str]:
    """Dotted engine tool names, e.g. ``{"bazi.chart", "time.now"}``."""
    source = "\n".join(
        path.read_text(encoding="utf-8") for path in sorted(CONTRACT_DIR.glob("*.go"))
    )
    return set(_METHOD_NAME.findall(source))


def engine_tool_names() -> set[str]:
    """MCP tool names: the registry dots become underscores."""
    return {name.replace(".", "_") for name in engine_method_names()}
