"""Operational HTTP routes for LIKI MCP."""

from __future__ import annotations

from app.config import load_config
from app.upstream import upstream_session
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse, Response
from starlette.routing import Route


async def healthz(_: Request) -> JSONResponse:
    return JSONResponse({"status": "ok"})


async def readyz(_: Request) -> JSONResponse:
    config = load_config()
    try:
        async with upstream_session(
            config.engine,
            client_name=config.server_name,
            client_version=config.server_version,
        ) as session:
            await session.ping()
        async with upstream_session(
            config.counsel,
            client_name=config.server_name,
            client_version=config.server_version,
        ) as session:
            await session.ping()
    except Exception as error:
        return JSONResponse({"status": "unavailable", "reason": str(error)}, status_code=503)
    return JSONResponse({"status": "ready"})


async def metrics(_: Request) -> Response:
    return Response(
        "# HELP liki_upstream_requests LIKI upstream requests.\n# TYPE liki_upstream_requests counter\n",
        media_type="text/plain; version=0.0.4",
    )


def create_app() -> Starlette:
    from app.gateway import mcp_endpoint

    return Starlette(
        routes=[
            Route("/mcp", mcp_endpoint, methods=["GET", "POST", "DELETE"]),
            Route("/healthz", healthz, methods=["GET"]),
            Route("/readyz", readyz, methods=["GET"]),
            Route("/metrics", metrics, methods=["GET"]),
        ]
    )
