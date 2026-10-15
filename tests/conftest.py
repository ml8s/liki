# Ensure contract-test helpers are importable under pytest.
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent))
# The counsel contract tests read the Pydantic models that own the tool schemas.
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "counsel"))


@pytest.fixture(autouse=True)
def _default_mcp_url(monkeypatch):
    # engine_mcp_client._endpoint() requires an explicit LIKI_MCP_URL; unit tests
    # mock urllib, so a placeholder URL is enough.
    monkeypatch.setenv("LIKI_MCP_URL", "http://engine.test.invalid/mcp")
