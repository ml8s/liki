"""Counsel aggregate MCP server.

The single Liki expert discovers every judgment capability through this
aggregate surface. Bazi/ziwei/liuyao/qimen/naming are internal technique
groups, not separate products or experts.
"""

import os
from contextlib import asynccontextmanager

import anyio
from starlette.applications import Starlette
from starlette.middleware import Middleware
from starlette.responses import JSONResponse, Response
from starlette.routing import Route

from app import engine_mcp_client  # noqa: E402
from app.counsel_servers import (  # noqa: E402
    RUNTIME_VERSION,
    create_counsel_root_mcp,
)
from app.http_middleware import (
    BodyLimitMiddleware,
    MetricsMiddleware,
)
from app.metrics import render_metrics  # noqa: E402


def _bind_host() -> str:
    return os.environ.get("LIKI_BIND_HOST", "127.0.0.1")


def _make_app():
    """Build the standard aggregate Starlette application."""
    root_app = create_counsel_root_mcp().streamable_http_app(
        streamable_http_path="/mcp",
        json_response=True,
        stateless_http=True,
        host="0.0.0.0",
    )

    async def healthz(_request):
        return JSONResponse({
            "status": "ok",
            "version": RUNTIME_VERSION,
        })

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
        return JSONResponse({
            "status": "ready",
            "version": RUNTIME_VERSION,
        })

    async def metrics(_request):
        return Response(
            render_metrics(),
            media_type="text/plain; version=0.0.4; charset=utf-8",
        )

    @asynccontextmanager
    async def lifespan(_app):
        if not os.environ.get("LIKI_ENGINE_MCP_URL", "").strip():
            raise RuntimeError(
                "LIKI_ENGINE_MCP_URL is required "
                "(engine-mcp endpoint, e.g. http://engine-mcp:8081/mcp)"
            )
        async with root_app.router.lifespan_context(root_app):
            yield

    return Starlette(
        routes=[
            Route("/healthz", healthz, methods=["GET"]),
            Route("/readyz", readyz, methods=["GET"]),
            Route("/metrics", metrics, methods=["GET"]),
            Route("/mcp", root_app),
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
