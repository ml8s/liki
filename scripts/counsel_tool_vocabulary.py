#!/usr/bin/env python3
"""Print the counsel tool vocabulary as JSON for the Go skill-document tests.

The counsel tool surface is owned by the Pydantic models in
``counsel/app/contracts``. The Go suite cannot import those models, so
``engine/internal/mcpcontract/skill_docs_sync_test.go`` shells out to this bridge
instead of reading a committed copy of the schemas: the models stay the single
source of truth for the fields the skill documents may reference.
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "counsel"))

from app.contracts import tool_vocabulary  # noqa: E402 - needs the path above


def main() -> int:
    json.dump(tool_vocabulary(), sys.stdout, ensure_ascii=False, sort_keys=True)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
