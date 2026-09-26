"""counsel MCP server（正交化判断层）。

暴露判断工具，按 domain 分域（bazi/ziwei/liuyao/qimen/naming）：
- bazi/ziwei：compute_factors / natal_query / period_query（因子快照 + 断语）
- naming：起名（surname/pick/char/compose/check，复用 qiming 逻辑 + engine 字库）
内部调 engine（排盘/字库）取判断所需字段，再在进程内求因子/断语/评估；
无状态、无存储依赖。counsel 是 engine 的补集（判断），排盘/历法在 engine。

Run:
    LIKI_MCP_URL=http://engine-mcp:8081/mcp \
        .venv/bin/python -m uvicorn app.counsel_mcp:app --host 127.0.0.1 --port 8086
"""
from __future__ import annotations

import collections
import copy
import hmac
import inspect
import json
import os
import pathlib
import sys
import time
from contextlib import AsyncExitStack, asynccontextmanager
from typing import Any

import anyio
from jsonschema import validators
from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError
from pydantic import BaseModel, ConfigDict, create_model
from starlette.applications import Starlette
from starlette.middleware import Middleware
from starlette.responses import JSONResponse
from starlette.routing import Route

from app import engine_client
from app.engine_client import MCPError as EngineMCPError
from app.natal.tools.errors import LikiToolError

JUDGMENT_SCHEMA = pathlib.Path(__file__).resolve().parent / "natal" / "tools" / "counsel-tools.json"
NAMING_SCHEMA = pathlib.Path(__file__).resolve().parent / "naming" / "tools" / "skill-tools.json"
NAMING_DIR = pathlib.Path(__file__).resolve().parent / "naming"
DIVINATION_SCHEMA = pathlib.Path(__file__).resolve().parent / "divination" / "tools" / "skill-tools.json"
DIVINATION_DIR = pathlib.Path(__file__).resolve().parent / "divination" / "tools"
VERSION_PATH = pathlib.Path(__file__).resolve().parent / "VERSION.txt"

RUNTIME_VERSION = VERSION_PATH.read_text(encoding="utf-8").strip()
_UNSET = object()


class _RawToolArguments(BaseModel):
    """Preserve undeclared arguments for validation against the manifest."""

    model_config = ConfigDict(extra="allow", arbitrary_types_allowed=True)

    def model_dump_one_level(self) -> dict[str, Any]:
        arguments = {
            field_info.alias or field_name: getattr(self, field_name)
            for field_name, field_info in type(self).model_fields.items()
        }
        arguments.update(self.__pydantic_extra__ or {})
        return arguments


def _raw_argument_model(tool_name: str, schema: dict) -> type[_RawToolArguments]:
    properties = schema.get("properties", {})
    required = set(schema.get("required", []))
    fields = {}
    for name in properties:
        default = properties[name].get("default", _UNSET)
        fields[name] = (Any, ... if name in required else default)
    return create_model(
        f"{tool_name}_RawArguments",
        __base__=_RawToolArguments,
        **fields,
    )

# natal 工具模块同目录 import（from duanyu / from paipan …），进程内调用需目录在 sys.path。
sys.path.insert(0, str(JUDGMENT_SCHEMA.parent))
sys.path.insert(0, str(NAMING_DIR))
sys.path.insert(0, str(DIVINATION_DIR))

from app.natal.tools import counsel  # noqa: E402


def _validate_arguments(schema: dict, arguments: dict) -> None:
    """在工具边界执行 manifest schema，避免 SDK 只按弱类型签名推断。"""
    validator_class = validators.validator_for(schema)
    validator_class.check_schema(schema)
    validator = validator_class(schema)
    errors = sorted(
        validator.iter_errors(arguments),
        key=lambda error: (
            ".".join(str(part) for part in error.absolute_path),
            error.message,
        ),
    )
    if not errors:
        return
    details = []
    for error in errors:
        path = ".".join(str(part) for part in error.absolute_path) or "$"
        details.append(f"{path}: {error.message}")
    raise ToolError("invalid_arguments: " + "; ".join(details)) from errors[0]


def _tool_signature(schema: dict) -> inspect.Signature:
    """给通用 handler 合成显式参数签名，供 MCP SDK 调用参数模型使用。"""
    required = set(schema.get("required", []))
    parameters: list[inspect.Parameter] = []
    for name in schema.get("properties", {}):
        if not name.isidentifier() or inspect.iskeyword(name):
            raise ValueError(f"MCP tool parameter is not a Python identifier: {name!r}")
        parameters.append(
            inspect.Parameter(
                name,
                inspect.Parameter.POSITIONAL_OR_KEYWORD,
                annotation=Any,
                default=None if name not in required else inspect.Parameter.empty,
            )
        )
    return inspect.Signature(parameters, return_annotation=str)


def _add_schema_tool(
    server: MCPServer,
    name: str,
    description: str,
    schema: dict,
    invoke,
) -> None:
    """注册工具并原样暴露 JSON schema。

    Python MCP SDK 的 ``add_tool`` 只接受 callable，并从函数签名推断 schema；
    对嵌套 JSON 契约会退化成 generic object/array。这里保留一个 ``**kwargs``
    handler 供 SDK 调用，在边界用 jsonschema 校验，再把已注册 Tool 的
    parameters 覆盖为 manifest 原文，确保 ``tools/list`` 不丢契约。
    """

    async def handler(**actual):
        arguments = {
            key: value for key, value in actual.items() if value is not _UNSET
        }
        if name == "qiming_compose":
            arguments.setdefault("second", [])
        _validate_arguments(schema, arguments)
        try:
            result = await anyio.to_thread.run_sync(invoke, arguments)
        except (ValueError, LikiToolError) as exc:
            raise ToolError(f"tool_failed: {exc}") from exc
        except EngineMCPError as exc:
            raise ToolError(f"engine_dependency_failed: {exc}") from exc
        return json.dumps(result, ensure_ascii=False)

    handler.__signature__ = _tool_signature(schema)

    server.add_tool(handler, name=name, description=description)
    registered = server._tool_manager.get_tool(name)  # noqa: SLF001 - SDK 未提供 schema setter
    if registered is None:
        raise RuntimeError(f"MCP tool was not registered: {name}")
    registered.fn_metadata.arg_model = _raw_argument_model(name, schema)
    registered.parameters = copy.deepcopy(schema)


def _require_domain(domain: str, chart: dict) -> None:
    """域校验：bazi 域只收八字盘，ziwei 域只收紫微盘。"""
    is_bazi = "ri" in chart
    if domain == "bazi" and not is_bazi:
        raise ToolError("invalid_arguments: counsel/bazi 只接收八字盘（chart 含四柱），收到紫微盘。")
    if domain == "ziwei" and is_bazi:
        raise ToolError("invalid_arguments: counsel/ziwei 只接收紫微盘（chart 为宫位结构），收到八字盘。")


def create_counsel_mcp(domain: str) -> MCPServer:
    if domain == "naming":
        return _create_naming_server()
    if domain in ("liuyao", "qimen"):
        return _create_divination_server(domain)

    if domain not in ("bazi", "ziwei"):
        raise ValueError(f"counsel domain 无效: {domain!r}")
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 判断层（{domain}）",
        description=(
            f"命理判断层（{domain}）：compute_factors 因子快照 + natal_query 本命断语 + "
            f"period_query 应期断语。排盘由 engine 完成（chart 为输入）。"
        ),
        version=RUNTIME_VERSION,
    )
    schema = json.loads(JUDGMENT_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        def invoke_natal(arguments: dict, _name=name):
            if _name == "compute_factors":
                _require_domain(domain, arguments["chart"])
                return counsel.compute_factors(arguments["chart"])
            if _name == "natal_query":
                return counsel.natal_query(
                    arguments["factors"],
                    arguments["topics"],
                    arguments["factors_digest"],
                    context=arguments.get("context"),
                    side=domain,
                )
            _require_domain(domain, arguments["chart"])
            return counsel.period_query(
                arguments["factors"],
                arguments["factors_digest"],
                arguments["time_scope"],
                arguments["topics"],
                arguments["chart"],
                side=domain,
            )

        _add_schema_tool(
            server,
            name=name,
            description=fn["description"],
            schema=fn["parameters"],
            invoke=invoke_natal,
        )
    return server


def create_counsel_root_mcp() -> MCPServer:
    """Create the aggregate counsel surface used by the root Liki skill.

    Bazi and Ziwei are intentionally excluded: those are exposed through the
    expert skills and domain-specific endpoints.
    """
    server = MCPServer(
        name="counsel",
        title="Liki Counsel",
        description="命理判断层：起名、六爻与奇门能力聚合。",
        version=RUNTIME_VERSION,
    )
    _add_naming_tools(server)
    _add_liuyao_tools(server)
    _add_qimen_tools(server)
    return server


def _add_divination_tools(server: MCPServer, domain: str) -> None:
    """六爻/奇门（liuyao/qimen）：snapshot 创建（排盘+因子）+ query 追问（ask→query）。

    复用 counsel 模块（create/ask——起卦/排盘/断语），engine 完成排盘；
    counsel 是补集（判断），ask 做成 query（追问出断语）。
    """
    if domain == "liuyao":
        from liuyao_ask import ask as query_ask
        from liuyao_snapshot import create as snapshot_create
    else:
        from qimen_ask import ask as query_ask
        from qimen_snapshot import create as snapshot_create
    schema = json.loads(DIVINATION_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        if name == f"{domain}_snapshot":
            impl = snapshot_create
            out_name = name
        elif name == f"{domain}_ask":
            impl = query_ask
            out_name = f"{domain}_query"  # ask → query（追问出断语）
        else:
            continue
        _add_schema_tool(
            server,
            name=out_name,
            description=fn["description"],
            schema=fn["parameters"],
            invoke=lambda arguments, _impl=impl: _impl(**arguments),
        )


def _create_divination_server(domain: str) -> MCPServer:
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 顾问（{domain}）",
        description=(
            f"{'六爻' if domain == 'liuyao' else '奇门'}顾问：snapshot 创建（起卦/排盘+因子），"
            f"query 追问（基于 snapshot 出断语，不重排）。排盘由 engine 完成。"
        ),
        version=RUNTIME_VERSION,
    )
    _add_divination_tools(server, domain)
    return server


def _qiming_check(given_names, yongshen, xishen=None, jishen=None):
    """适配 schema 参数名（yongshen/xishen/jishen）→ evaluate_names（yong_shen 等）。"""
    from qiming import evaluate_names

    return evaluate_names(
        given_names, yongshen, xishen or [], jishen or []
    )


def _add_naming_tools(server: MCPServer) -> None:
    """起名（naming）：qiming 工具复用 qiming 逻辑（字库来自 engine）。"""
    from qiming import compose_names, lookup_char, match_surnames, pick_chars

    qiming_fns = {
        "qiming_surname": match_surnames,
        "qiming_pick": pick_chars,
        "qiming_char": lookup_char,
        "qiming_compose": compose_names,
        "qiming_check": _qiming_check,
    }
    schema = json.loads(NAMING_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        _add_schema_tool(
            server,
            name=name,
            description=fn["description"],
            schema=fn["parameters"],
            invoke=lambda arguments, _name=name: qiming_fns[_name](**arguments),
        )


def _add_liuyao_tools(server: MCPServer) -> None:
    _add_divination_tools(server, "liuyao")


def _add_qimen_tools(server: MCPServer) -> None:
    _add_divination_tools(server, "qimen")


def _create_naming_server() -> MCPServer:
    server = MCPServer(
        name="counsel-naming",
        title="Liki 起名顾问",
        description=(
            "起名顾问：姓氏匹配、按五行取字、单字查询、组名、候选名评估。"
            "字库来自 engine（继承），用神/喜忌由调用方传入（八字判断结果）。"
        ),
        version=RUNTIME_VERSION,
    )
    _add_naming_tools(server)
    return server


def _make_app():
    """Build the standard Starlette application for all Counsel domains.

    Each domain is a complete MCP Streamable HTTP application generated by the
    official MCP SDK. Starlette owns routing and the combined ASGI lifespan.
    The public `/counsel` prefix is stripped by the gateway; this process owns
    only `/mcp` and `/mcp/{domain}`.
    """
    domains = ("bazi", "ziwei", "liuyao", "qimen", "naming")
    root_app = create_counsel_root_mcp().streamable_http_app(
        streamable_http_path="/mcp",
        json_response=True,
        stateless_http=True,
        host="0.0.0.0",
    )
    domain_apps = {
        domain: create_counsel_mcp(domain).streamable_http_app(
            streamable_http_path=f"/mcp/{domain}",
            json_response=True,
            stateless_http=True,
            host="0.0.0.0",
        )
        for domain in domains
    }

    async def healthz(_request):
        return JSONResponse({
            "status": "ok",
            "version": RUNTIME_VERSION,
            "domains": list(domains),
        })

    async def readyz(_request):
        try:
            engine_version = await anyio.to_thread.run_sync(engine_client.engine_version)
            await anyio.to_thread.run_sync(
                engine_client.ensure_engine_compatible,
                engine_version,
                engine_client.required_engine_version(),
            )
        except Exception:
            return JSONResponse(
                {"status": "unready", "reason": "engine dependency unavailable"},
                status_code=503,
            )
        return JSONResponse({
            "status": "ready",
            "version": RUNTIME_VERSION,
            "engine_version": engine_version,
        })

    @asynccontextmanager
    async def lifespan(app: Starlette):
        """Run each official MCP domain app's lifespan exactly once."""
        async with AsyncExitStack() as stack:
            await stack.enter_async_context(root_app.router.lifespan_context(root_app))
            for domain in domains:
                domain_app = domain_apps[domain]
                await stack.enter_async_context(
                    domain_app.router.lifespan_context(domain_app)
                )
            yield

    class RateLimiter:
        """Small non-durable limiter; enough for one-process MCP service."""

        def __init__(self, limit: int, window_seconds: float, max_keys: int = 65536):
            self.limit = limit
            self.window_seconds = window_seconds
            self.max_keys = max_keys
            self.hits: dict[str, collections.deque[float]] = {}
            self._last_cleanup = 0.0

        def allow(self, key: str) -> bool:
            now = time.monotonic()
            if now - self._last_cleanup >= self.window_seconds:
                self._cleanup(now)
            hits = self.hits.get(key)
            if hits is None:
                if len(self.hits) >= self.max_keys:
                    self._evict_oldest()
                hits = collections.deque()
                self.hits[key] = hits
            while hits and hits[0] <= now - self.window_seconds:
                hits.popleft()
            if len(hits) >= self.limit:
                return False
            hits.append(now)
            return True

        def _cleanup(self, now: float) -> None:
            self._last_cleanup = now
            for key, hits in list(self.hits.items()):
                while hits and hits[0] <= now - self.window_seconds:
                    hits.popleft()
                if not hits:
                    self.hits.pop(key, None)

        def _evict_oldest(self) -> None:
            oldest_key = None
            oldest_timestamp = None
            for key, hits in self.hits.items():
                timestamp = hits[0] if hits else float("-inf")
                if oldest_timestamp is None or timestamp < oldest_timestamp:
                    oldest_key = key
                    oldest_timestamp = timestamp
            if oldest_key is not None:
                self.hits.pop(oldest_key, None)

    class AuthMiddleware:
        """Optional bearer-token gate for self-hosted MCP deployments."""

        def __init__(self, app, token: str):
            self.app = app
            self.token = token

        async def __call__(self, scope, receive, send):
            if scope["type"] != "http":
                await self.app(scope, receive, send)
                return
            if (
                not self.token
                or scope.get("path") in ("/healthz", "/readyz")
                or scope.get("method") == "OPTIONS"
            ):
                await self.app(scope, receive, send)
                return

            authorization = ""
            for name, value in scope.get("headers") or []:
                if name == b"authorization":
                    authorization = value.decode("latin-1")
                    break
            expected = f"Bearer {self.token}"
            if hmac.compare_digest(authorization, expected):
                await self.app(scope, receive, send)
                return

            response = JSONResponse(
                {
                    "error": {
                        "code": "unauthorized",
                        "message": "missing or invalid bearer token",
                    }
                },
                status_code=401,
                headers={"WWW-Authenticate": "Bearer"},
            )
            await response(scope, receive, send)

    class RateLimitMiddleware:
        """Transparent ASGI middleware for this single-process service."""

        def __init__(
            self,
            app,
            limit: int,
            window_seconds: float,
            max_keys: int,
        ):
            self.app = app
            self.rate_limiter = RateLimiter(limit, window_seconds, max_keys)

        def client_key(self, scope: dict) -> str:
            try:
                trusted_hops = int(os.environ.get("LIKI_TRUSTED_PROXY_HOPS", "0"))
            except ValueError:
                trusted_hops = 0
            if trusted_hops > 0:
                for name, value in scope.get("headers") or []:
                    if name == b"x-forwarded-for":
                        parts = [item.strip() for item in value.decode("latin-1").split(",")]
                        if not parts or parts[-1] == "":
                            break
                        index = max(len(parts) - trusted_hops, 0)
                        return parts[index]
            client = scope.get("client") or ("unknown", 0)
            return str(client[0])

        async def __call__(self, scope, receive, send):
            if scope["type"] != "http":
                await self.app(scope, receive, send)
                return
            if scope.get("path") in ("/healthz", "/readyz"):
                await self.app(scope, receive, send)
                return
            if not self.rate_limiter.allow(self.client_key(scope)):
                response = JSONResponse(
                    {"error": {"code": "rate_limited", "message": "too many requests"}},
                    status_code=429,
                )
                await response(scope, receive, send)
                return
            await self.app(scope, receive, send)

    class BodyLimitMiddleware:
        """Reject oversized MCP JSON payloads before business logic runs."""

        def __init__(self, app, max_body_bytes: int):
            self.app = app
            self.max_body_bytes = max_body_bytes

        def content_length(self, scope: dict) -> int | None:
            for name, value in scope.get("headers") or []:
                if name != b"content-length":
                    continue
                try:
                    return int(value.decode("latin-1"))
                except ValueError:
                    return None
            return None

        async def __call__(self, scope, receive, send):
            if scope["type"] != "http":
                await self.app(scope, receive, send)
                return
            if (length := self.content_length(scope)) is not None:
                if length > self.max_body_bytes:
                    response = JSONResponse(
                        {"error": {"code": "payload_too_large", "message": "request body exceeds the configured limit"}},
                        status_code=413,
                    )
                    await response(scope, receive, send)
                    return
                await self.app(scope, receive, send)
                return

            max_body_bytes = self.max_body_bytes
            messages = []
            received = 0

            while True:
                message = await receive()
                messages.append(message)
                if message.get("type") == "http.request":
                    received += len(message.get("body", b""))
                    if received > max_body_bytes:
                        response = JSONResponse(
                            {"error": {"code": "payload_too_large", "message": "request body exceeds the configured limit"}},
                            status_code=413,
                        )
                        await response(scope, receive, send)
                        return
                if message.get("type") == "http.disconnect" or not message.get("more_body", False):
                    break

            replay_index = 0

            async def replay_receive():
                nonlocal replay_index
                if replay_index >= len(messages):
                    return {"type": "http.disconnect"}
                message = messages[replay_index]
                replay_index += 1
                return message

            await self.app(scope, replay_receive, send)

    routes = [
        Route("/healthz", healthz, methods=["GET"]),
        Route("/readyz", readyz, methods=["GET"]),
        *(
            Route(f"/mcp/{domain}", endpoint=domain_apps[domain])
            for domain in domains
        ),
        Route("/mcp", endpoint=root_app),
    ]
    return Starlette(
        routes=routes,
        lifespan=lifespan,
        middleware=[
            Middleware(
                RateLimitMiddleware,
                limit=int(os.environ.get("LIKI_COUNSEL_RATE_LIMIT", "240")),
                window_seconds=float(
                    os.environ.get("LIKI_COUNSEL_RATE_WINDOW_SECONDS", "60")
                ),
                max_keys=int(
                    os.environ.get("LIKI_COUNSEL_RATE_MAX_KEYS", "65536")
                ),
            ),
            Middleware(
                AuthMiddleware,
                token=os.environ.get("LIKI_MCP_TOKEN", ""),
            ),
            Middleware(
                BodyLimitMiddleware,
                max_body_bytes=int(os.environ.get("LIKI_MCP_MAX_BODY_BYTES", str(1024 * 1024))),
            ),
        ],
    )

app = _make_app()

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8091)
