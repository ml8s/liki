"""Release identity and retired-name migration contract."""

import asyncio

from app.gateway import create_app
from httpx2 import ASGITransport, AsyncClient

from test_gateway import _call


def test_healthz_has_no_release_contract(monkeypatch):
    async def run():
        monkeypatch.setenv("LIKI_GIT_SHA", "d2aaecbd0557")
        app = create_app()
        async with AsyncClient(
            transport=ASGITransport(app=app), base_url="http://test"
        ) as client:
            return await client.get("/healthz")

    response = asyncio.run(run())
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_retired_tools_fail_with_migration_hint():
    cases = {
        "compute_factors": "removed; factors are internal; use natal_query",
        "liuyao_snapshot": "renamed; use liuyao_ask",
        "period_query": "renamed; use natal_liunian_query",
        "qimen_snapshot": "renamed; use qimen_ask",
        "tianwen_resolve": "renamed; use subject_candidates",
    }
    for name, replacement in cases.items():
        response = _call({
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {"name": name, "arguments": {}},
        })
        assert response.status_code == 200
        error = response.json()["error"]
        assert error["code"] == -32602
        assert error["message"] == f"unknown tool: {name}; {replacement}"
