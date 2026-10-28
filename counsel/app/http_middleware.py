"""HTTP middleware shared by the Counsel MCP application."""


import hmac
import time

from starlette.responses import JSONResponse


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
            or scope.get("path") in ("/healthz", "/readyz", "/metrics")
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


class MetricsMiddleware:
    """Record fixed-template HTTP route, method, and status-class metrics."""

    def __init__(self, app):
        self.app = app

    async def __call__(self, scope, receive, send):
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        started = time.perf_counter()
        status_class = "unknown"

        async def send_wrapper(message):
            nonlocal status_class
            if message.get("type") == "http.response.start":
                status = int(message.get("status", 500))
                status_class = f"{status // 100}xx"
            await send(message)

        try:
            await self.app(scope, receive, send_wrapper)
        finally:
            from app.metrics import observe_http_request

            observe_http_request(
                route=http_route(scope.get("path", "")),
                method=scope.get("method", ""),
                status_class=status_class,
                duration=time.perf_counter() - started,
            )


def http_route(path: str) -> str:
    known = {"/healthz", "/readyz", "/metrics", "/mcp"}
    if path in known:
        return path
    if path.startswith("/mcp/"):
        return "/mcp/{domain}"
    return "unmatched"


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

    async def _payload_too_large(self, scope, receive, send) -> None:
        response = JSONResponse(
            {"error": {"code": "payload_too_large", "message": "request body exceeds the configured limit"}},
            status_code=413,
        )
        await response(scope, receive, send)

    async def __call__(self, scope, receive, send):
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        declared_length = self.content_length(scope)
        if declared_length is not None and declared_length > self.max_body_bytes:
            await self._payload_too_large(scope, receive, send)
            return

        messages = []
        received = 0
        while True:
            message = await receive()
            messages.append(message)
            if message.get("type") == "http.request":
                received += len(message.get("body", b""))
                if received > self.max_body_bytes:
                    await self._payload_too_large(scope, receive, send)
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
