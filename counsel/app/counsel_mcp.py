"""Counsel aggregate MCP server.

The single Liki expert discovers every judgment capability through this
aggregate surface. Bazi/ziwei/liuyao/qimen/naming are internal technique
groups, not separate products or experts.
"""

import logging
import os
from contextlib import asynccontextmanager

import anyio
from starlette.applications import Starlette
from starlette.middleware import Middleware
from starlette.responses import JSONResponse, Response
from starlette.routing import Route

from app import engine_mcp_client  # noqa: E402
from app.counsel_servers import create_counsel_root_mcp  # noqa: E402
from app.http_middleware import (
    BodyLimitMiddleware,
    MetricsMiddleware,
)
from app.metrics import render_metrics  # noqa: E402

logger = logging.getLogger(__name__)


def _bind_host() -> str:
    return os.environ.get("LIKI_BIND_HOST", "127.0.0.1")


def _prewarm_caches() -> None:
    """Eagerly load read-only data tables so the first request skips lazy init.

    Best-effort: a data table with a validation error logs a warning at startup
    instead of crashing the server for all domains. The lazy loader still
    raises at request time for the affected tool.
    """
    try:
        from app.divination.tools.divination_safety import _load_rules  # noqa: SLF001
        from app.divination.tools.liuyao_conditions import load_rules
        from app.divination.tools.liuyao_matters import load_matter_table
        from app.divination.tools.qimen_interpretations import (
            load_interpretation_index,
            load_rule_table,
        )
        from app.divination.tools.qimen_matters import load_matter_table as _load_qimen_matters
        from app.natal.tools.factor_constants import load_constants
        from app.natal.tools.factor_tables import load_factor_rows, load_liunian_rows
        from app.natal.tools.natal_projection import load_contract
        from app.natal.tools.yongshen import load_policy

        load_constants()
        load_factor_rows()
        load_liunian_rows()
        load_policy()
        load_contract()
        load_rules()
        load_matter_table()
        _load_qimen_matters()
        load_rule_table()
        load_interpretation_index()
        _load_rules()
    except Exception:
        logger.warning(
            "cache prewarm failed; affected tools will lazy-load on first request",
            exc_info=True,
        )


def _make_app():
    """Build the standard aggregate Starlette application."""
    root_app = create_counsel_root_mcp().streamable_http_app(
        streamable_http_path="/mcp",
        json_response=True,
        stateless_http=True,
        host="0.0.0.0",
    )

    async def healthz(_request):
        return JSONResponse({"status": "ok"})

    async def readyz(_request):
        try:
            timeout = float(os.environ.get("LIKI_READYZ_TIMEOUT_SECONDS", "1"))

            def discover_engine():
                return engine_mcp_client.server_info(
                    retries=0,
                    timeout=timeout,
                )

            await anyio.to_thread.run_sync(discover_engine)
        except (OSError, TimeoutError, ValueError, engine_mcp_client.MCPError):
            return JSONResponse(
                {
                    "status": "unavailable",
                    "reason": "engine_dependency_unreachable",
                },
                status_code=503,
            )
        return JSONResponse({"status": "ready"})

    async def metrics(_request):
        return Response(
            render_metrics(),
            media_type="text/plain; version=0.0.4; charset=utf-8",
        )

    class MCPRoute:
        """ASGI route that rejects unneeded listener streams at the boundary."""

        async def __call__(self, scope, receive, send):
            method = next(
                (
                    value.decode("latin-1")
                    for name, value in scope.get("headers") or []
                    if name == b"mcp-method"
                ),
                "",
            )
            if method == "subscriptions/listen":
                await JSONResponse(
                    {
                        "jsonrpc": "2.0",
                        "id": None,
                        "error": {
                            "code": -32601,
                            "message": "method not found: subscriptions/listen",
                        },
                    },
                    status_code=404,
                )(scope, receive, send)
                return
            await root_app(scope, receive, send)

    mcp_endpoint = MCPRoute()

    @asynccontextmanager
    async def lifespan(_app):
        if not os.environ.get("LIKI_ENGINE_MCP_URL", "").strip():
            raise RuntimeError(
                "LIKI_ENGINE_MCP_URL is required "
                "(engine-mcp endpoint, e.g. http://engine-mcp:8081/mcp)"
            )
        await anyio.to_thread.run_sync(_prewarm_caches)
        async with root_app.router.lifespan_context(root_app):
            yield

    return Starlette(
        routes=[
            Route("/healthz", healthz, methods=["GET"]),
            Route("/readyz", readyz, methods=["GET"]),
            Route("/metrics", metrics, methods=["GET"]),
            # MCP 2026 has no standalone GET stream. Keep the SDK's POST
            # endpoint from accepting legacy GET probes that would open an
            # unusable stateless SSE response.
            Route("/mcp", mcp_endpoint, methods=["POST"]),
        ],
        lifespan=lifespan,
        middleware=[
            Middleware(MetricsMiddleware),
            Middleware(
                BodyLimitMiddleware,
                max_body_bytes=int(
                    os.environ.get(
                        "LIKI_HTTP_MAX_BODY_BYTES",
                        str(1024 * 1024),
                    )
                ),
            ),
        ],
    )


app = _make_app()

if __name__ == "__main__":
    import uvicorn

    os.environ.setdefault("LIKI_BIND_HOST", "127.0.0.1")
    uvicorn.run(app, host=_bind_host(), port=8086)
