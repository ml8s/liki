# Ensure contract-test helpers are importable under pytest.
from pathlib import Path
import sys

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parent))


@pytest.fixture(autouse=True)
def _default_mcp_url(monkeypatch):
    # engine_client._endpoint() requires an explicit LIKI_MCP_URL; unit tests
    # mock urllib, so a placeholder URL is enough.
    monkeypatch.setenv("LIKI_MCP_URL", "http://engine.test.invalid/mcp")
