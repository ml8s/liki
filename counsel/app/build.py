"""Immutable build identity shared by Liki MCP runtime surfaces."""

from __future__ import annotations

import os


def build_id() -> str:
    """Return the Git short SHA injected at image build time."""
    return os.environ.get("LIKI_GIT_SHA", "").strip() or "unknown"
