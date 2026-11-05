"""MCP client adapter for LIKI's internal upstream servers."""

from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncIterator

from httpx2 import AsyncClient
from mcp import ClientSession
from mcp.client.streamable_http import streamable_http_client
from mcp.types import Implementation

from app.config import UpstreamConfig


class UpstreamError(RuntimeError):
    """An internal MCP upstream request failed."""


@asynccontextmanager
async def upstream_session(
    config: UpstreamConfig,
    *,
    client_name: str = "liki-mcp",
    client_version: str = "0.1.0",
) -> AsyncIterator[ClientSession]:
    """Open a short-lived MCP session against one internal upstream."""
    async with AsyncClient(timeout=config.timeout_seconds) as http_client:
        async with streamable_http_client(
            config.url,
            http_client=http_client,
            terminate_on_close=True,
        ) as (read_stream, write_stream):
            async with ClientSession(
                read_stream,
                write_stream,
                read_timeout_seconds=config.timeout_seconds,
                client_info=Implementation(
                    name=client_name,
                    version=client_version,
                ),
            ) as session:
                await session.initialize()
                yield session
